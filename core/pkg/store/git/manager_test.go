/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import (
	"archive/tar"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/suite"
)

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

type ManagerTestSuite struct {
	suite.Suite
	mgr     Manager
	rootDir string
}

func (s *ManagerTestSuite) SetupTest() {
	s.rootDir = s.T().TempDir()
	mgr, err := New(s.rootDir)
	s.Require().NoError(err)
	s.mgr = mgr
}

func (s *ManagerTestSuite) TearDownTest() {
	if s.mgr != nil {
		_ = s.mgr.Close()
	}
}

// --- Helper Methods ---

type testGitRepoInfo struct {
	RepoDir        string
	MainCommitHash string
	FeatureCommit  string
	TagName        string
	BranchName     string
}

func (s *ManagerTestSuite) initTestGitRepo(files, symlinks map[string]string) *testGitRepoInfo {
	s.T().Helper()
	repoDir := s.T().TempDir()

	gitRepo, err := git.PlainInit(repoDir, false)
	s.Require().NoError(err)

	wt, err := gitRepo.Worktree()
	s.Require().NoError(err)

	// Write regular files
	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		s.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o755))
		s.Require().NoError(os.WriteFile(fullPath, []byte(content), 0o644))
		_, err = wt.Add(relPath)
		s.Require().NoError(err)
	}

	// Write symlinks
	for linkName, target := range symlinks {
		fullPath := filepath.Join(repoDir, linkName)
		s.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o755))
		s.Require().NoError(os.Symlink(target, fullPath))
		_, err = wt.Add(linkName)
		s.Require().NoError(err)
	}

	// Commit on default branch
	sig := &object.Signature{
		Name:  "Drassi Tester",
		Email: "tester@drassi.run",
		When:  time.Now(),
	}
	mainHash, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	s.Require().NoError(err)

	// Create a tag
	tagName := "v1.0.0"
	_, err = gitRepo.CreateTag(tagName, mainHash, &git.CreateTagOptions{
		Tagger:  sig,
		Message: "Release v1.0.0",
	})
	s.Require().NoError(err)

	// Create and checkout a feature branch
	branchName := "feature-test"
	headRef, err := gitRepo.Head()
	s.Require().NoError(err)

	err = wt.Checkout(&git.CheckoutOptions{
		Hash:   headRef.Hash(),
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})
	s.Require().NoError(err)

	// Add a file on the feature branch
	featureFilePath := filepath.Join(repoDir, "feature.txt")
	s.Require().NoError(os.WriteFile(featureFilePath, []byte("feature branch content"), 0o644))
	_, err = wt.Add("feature.txt")
	s.Require().NoError(err)

	featHash, err := wt.Commit("Feature commit", &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	s.Require().NoError(err)

	// Checkout back to main
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: headRef.Name(),
	})
	s.Require().NoError(err)

	return &testGitRepoInfo{
		RepoDir:        repoDir,
		MainCommitHash: mainHash.String(),
		FeatureCommit:  featHash.String(),
		TagName:        tagName,
		BranchName:     branchName,
	}
}

func (s *ManagerTestSuite) assertTarEntries(r io.Reader, expectedEntries, expectedSymlinks map[string]string) {
	s.T().Helper()
	entries := make(map[string]string)
	symlinks := make(map[string]string)
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		s.Require().NoError(err)

		if hdr.Typeflag == tar.TypeSymlink {
			symlinks[hdr.Name] = hdr.Linkname
		} else if hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			s.Require().NoError(err)
			entries[hdr.Name] = string(content)
		}
	}

	if expectedEntries != nil {
		s.Assert().Equal(expectedEntries, entries)
	} else {
		s.Assert().Empty(entries)
	}

	if expectedSymlinks != nil {
		s.Assert().Equal(expectedSymlinks, symlinks)
	} else {
		s.Assert().Empty(symlinks)
	}
}

func (s *ManagerTestSuite) makeLocalRepoRef(repoDir, ref string) *RepoReference {
	trimmed := strings.TrimPrefix(repoDir, "/")
	return &RepoReference{
		Scheme:    "git",
		Transport: "file",
		Endpoint:  "localhost",
		Name:      trimmed,
		Ref:       ref,
	}
}

func (s *ManagerTestSuite) newTestManager() (Manager, string) {
	s.T().Helper()
	rootDir := s.T().TempDir()
	mgr, err := New(rootDir)
	s.Require().NoError(err)
	return mgr, rootDir
}

// --- Tests ---

func (s *ManagerTestSuite) TestNew() {
	s.Run("success", func() {
		tempDir := s.T().TempDir()
		storeDir := filepath.Join(tempDir, "store")
		mgr, err := New(storeDir)
		s.Require().NoError(err)
		s.Require().NotNil(mgr)

		info, err := os.Stat(storeDir)
		s.Require().NoError(err)
		s.Assert().True(info.IsDir())
	})

	s.Run("invalid dir error", func() {
		tempDir := s.T().TempDir()
		filePath := filepath.Join(tempDir, "existing-file")
		s.Require().NoError(os.WriteFile(filePath, []byte("data"), 0o644))

		// Attempting to create manager where root is a child of a file
		invalidPath := filepath.Join(filePath, "child")
		mgr, err := New(invalidPath)
		s.Assert().Error(err)
		s.Assert().Nil(mgr)
	})
}

func (s *ManagerTestSuite) TestFetch() {
	files := map[string]string{
		"action.yml": "name: test action",
		"index.js":   "console.log('hello')",
	}
	symlinks := map[string]string{
		"link.js": "index.js",
	}
	repoInfo := s.initTestGitRepo(files, symlinks)

	s.Run("fetch default branch (HEAD)", func() {
		ref := s.makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

		rev, err := s.mgr.Fetch(s.T().Context(), ref)
		s.Require().NoError(err)
		s.Assert().Equal(repoInfo.MainCommitHash, rev)
	})

	s.Run("fetch feature branch", func() {
		ref := s.makeLocalRepoRef(repoInfo.RepoDir, repoInfo.BranchName)

		rev, err := s.mgr.Fetch(s.T().Context(), ref)
		s.Require().NoError(err)
		s.Assert().Equal(repoInfo.FeatureCommit, rev)
	})

	s.Run("fetch tag", func() {
		ref := s.makeLocalRepoRef(repoInfo.RepoDir, repoInfo.TagName)

		rev, err := s.mgr.Fetch(s.T().Context(), ref)
		s.Require().NoError(err)
		s.Assert().Equal(repoInfo.MainCommitHash, rev)
	})

	s.Run("fetch non-existent ref returns error", func() {
		ref := s.makeLocalRepoRef(repoInfo.RepoDir, "nonexistent-branch-404")

		rev, err := s.mgr.Fetch(s.T().Context(), ref)
		s.Assert().Error(err)
		s.Assert().Empty(rev)
	})

	s.Run("fetch with auth token propagates basic auth", func() {
		var receivedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		ref := &RepoReference{
			Transport: "http",
			Endpoint:  strings.TrimPrefix(server.URL, "http://"),
			Name:      "test/repo",
			Ref:       "HEAD",
		}

		const testToken = "secret-token-123"
		_, err := s.mgr.Fetch(s.T().Context(), ref, WithToken(testToken))
		s.Assert().Error(err)
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("token:"+testToken))
		s.Assert().Equal(expectedAuth, receivedAuth)
	})

	s.Run("fetch non-existent repository returns error", func() {
		ref := s.makeLocalRepoRef(filepath.Join(s.T().TempDir(), "does-not-exist"), "HEAD")

		rev, err := s.mgr.Fetch(s.T().Context(), ref)
		s.Assert().Error(err)
		s.Assert().Empty(rev)
	})
}

func (s *ManagerTestSuite) TestRead() {
	files := map[string]string{
		"action.yml":           "name: test-action\ndescription: test",
		"src/index.js":         "console.log('test');",
		"nested/dir/file.md":   "# Docs",
		"nested/sub/data.json": `{"key": "value"}`,
	}
	symlinks := map[string]string{
		"src/alias.js": "index.js",
	}
	repoInfo := s.initTestGitRepo(files, symlinks)

	ref := s.makeLocalRepoRef(repoInfo.RepoDir, "HEAD")
	rev, err := s.mgr.Fetch(s.T().Context(), ref)
	s.Require().NoError(err)

	s.Run("read full archive with directory prefix", func() {
		prefix := "actions/my-action@v1"
		rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithSubpath(prefix))
		s.Require().NoError(err)
		defer rc.Close()

		expectedEntries := map[string]string{
			"actions/my-action@v1/action.yml":           files["action.yml"],
			"actions/my-action@v1/src/index.js":         files["src/index.js"],
			"actions/my-action@v1/nested/dir/file.md":   files["nested/dir/file.md"],
			"actions/my-action@v1/nested/sub/data.json": files["nested/sub/data.json"],
		}
		expectedSymlinks := map[string]string{
			"actions/my-action@v1/src/alias.js": "index.js",
		}

		s.assertTarEntries(rc, expectedEntries, expectedSymlinks)
	})

	s.Run("read full archive with empty prefix", func() {
		rc, err := s.mgr.Read(s.T().Context(), ref, rev)
		s.Require().NoError(err)
		defer rc.Close()

		s.assertTarEntries(rc, files, symlinks)
	})

	s.Run("read non-existent revision error", func() {
		nonExistentRev := strings.Repeat("0", 40)
		rc, err := s.mgr.Read(s.T().Context(), ref, nonExistentRev)
		s.Assert().Error(err)
		s.Assert().Nil(rc)
	})

	s.Run("read non-existent repo error", func() {
		unknownRef := &RepoReference{
			Endpoint: "unknown.com",
			Name:     "unknown/repo",
			Ref:      "main",
		}
		rc, err := s.mgr.Read(s.T().Context(), unknownRef, rev)
		s.Assert().Error(err)
		s.Assert().Nil(rc)
	})

	s.Run("read existing regular file", func() {
		rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithFile("action.yml"))
		s.Require().NoError(err)
		defer rc.Close()

		content, err := io.ReadAll(rc)
		s.Require().NoError(err)
		s.Assert().Equal(files["action.yml"], string(content))
	})

	s.Run("read nested file with leading slash and relative dots", func() {
		testPaths := []string{
			"nested/sub/data.json",
			"/nested/sub/data.json",
			"./nested/sub/data.json",
			"nested/../nested/sub/data.json",
		}
		for _, p := range testPaths {
			rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithFile(p))
			s.Require().NoError(err, "path: %s", p)
			content, err := io.ReadAll(rc)
			s.Require().NoError(err)
			s.Assert().Equal(files["nested/sub/data.json"], string(content))
			_ = rc.Close()
		}
	})

	s.Run("read non-existent file error", func() {
		rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithFile("does-not-exist.txt"))
		s.Assert().Error(err)
		s.Assert().Nil(rc)
		s.Assert().True(os.IsNotExist(err) || notFoundErr(err))
	})

	s.Run("read directory as file error", func() {
		rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithFile("src"))
		s.Assert().Error(err)
		s.Assert().Nil(rc)
		s.Assert().Contains(err.Error(), "not a (regular) file")
	})

	s.Run("read file with invalid revision error", func() {
		invalidRev := strings.Repeat("a", 40)
		rc, err := s.mgr.Read(s.T().Context(), ref, invalidRev, WithFile("action.yml"))
		s.Assert().Error(err)
		s.Assert().Nil(rc)
	})

	s.Run("read file from non-existent repo error", func() {
		unknownRef := &RepoReference{
			Endpoint: "unknown.com",
			Name:     "unknown/repo",
			Ref:      "main",
		}
		rc, err := s.mgr.Read(s.T().Context(), unknownRef, rev, WithFile("action.yml"))
		s.Assert().Error(err)
		s.Assert().Nil(rc)
	})

	s.Run("read with both file and subpath returns error", func() {
		rc, err := s.mgr.Read(s.T().Context(), ref, rev, WithFile("action.yml"), WithSubpath("actions"))
		s.Assert().Error(err)
		s.Assert().Nil(rc)
		s.Assert().Contains(err.Error(), "cannot specify both file and subpath")
	})

	s.Run("context cancellation during read", func() {
		cancelMgr, _ := s.newTestManager()

		cancelRev, err := cancelMgr.Fetch(s.T().Context(), ref)
		s.Require().NoError(err)

		ctx, cancel := context.WithCancel(s.T().Context())
		cancel()

		rc, err := cancelMgr.Read(ctx, ref, cancelRev)
		s.Require().NoError(err)
		defer rc.Close()

		_, readErr := io.ReadAll(rc)
		s.Assert().Error(readErr)
		s.Assert().True(errors.Is(readErr, context.Canceled) || errors.Is(readErr, io.ErrClosedPipe))
	})
}

func (s *ManagerTestSuite) TestPersistence() {
	files := map[string]string{
		"action.yml": "name: persistent-action",
	}
	repoInfo := s.initTestGitRepo(files, nil)

	rootDir := s.T().TempDir()

	// Manager 1: Fetches the repo
	mgr1, err := New(rootDir)
	s.Require().NoError(err)
	ref := s.makeLocalRepoRef(repoInfo.RepoDir, "HEAD")
	rev, err := mgr1.Fetch(s.T().Context(), ref)
	s.Require().NoError(err)

	// Manager 2: Created with the same rootDir without calling Fetch
	mgr2, err := New(rootDir)
	s.Require().NoError(err)

	// File lookup should succeed by reading from disk
	rc, err := mgr2.Read(s.T().Context(), ref, rev, WithFile("action.yml"))
	s.Require().NoError(err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	s.Require().NoError(err)
	s.Assert().Equal(files["action.yml"], string(content))

	// Read archive should also succeed
	tarRc, err := mgr2.Read(s.T().Context(), ref, rev)
	s.Require().NoError(err)
	defer tarRc.Close()
	s.assertTarEntries(tarRc, files, nil)
}

func (s *ManagerTestSuite) TestConcurrency() {
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	repoInfo := s.initTestGitRepo(files, nil)

	ref := s.makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := s.T().Context()

			// Concurrently fetch
			rev, err := s.mgr.Fetch(ctx, ref)
			s.Require().NoError(err)

			// Concurrently read file
			rc, err := s.mgr.Read(ctx, ref, rev, WithFile("file1.txt"))
			s.Require().NoError(err)
			content, err := io.ReadAll(rc)
			s.Require().NoError(err)
			s.Assert().Equal("content1", string(content))
			_ = rc.Close()

			// Concurrently read tar
			tarRc, err := s.mgr.Read(ctx, ref, rev)
			s.Require().NoError(err)
			s.assertTarEntries(tarRc, files, nil)
			_ = tarRc.Close()
		}(i)
	}

	wg.Wait()
}

func (s *ManagerTestSuite) TestClose() {
	files := map[string]string{
		"action.yml": "name: close-test",
	}
	repoInfo := s.initTestGitRepo(files, nil)

	ref := s.makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

	// Fetch repository to ensure it is opened and stored in manager
	_, err := s.mgr.Fetch(s.T().Context(), ref)
	s.Require().NoError(err)

	// Close the manager
	err = s.mgr.Close()
	s.Require().NoError(err)

	// Calling Close again should also succeed without error
	err = s.mgr.Close()
	s.Require().NoError(err)
}
