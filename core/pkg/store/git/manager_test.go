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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper Functions ---

func newTestManager(t *testing.T) (Manager, string) {
	t.Helper()
	rootDir := t.TempDir()
	mgr, err := New(rootDir)
	require.NoError(t, err)
	return mgr, rootDir
}

type testGitRepoInfo struct {
	RepoDir        string
	MainCommitHash string
	FeatureCommit  string
	TagName        string
	BranchName     string
}

func initTestGitRepo(t *testing.T, files, symlinks map[string]string) *testGitRepoInfo {
	t.Helper()
	repoDir := t.TempDir()

	gitRepo, err := git.PlainInit(repoDir, false)
	require.NoError(t, err)

	wt, err := gitRepo.Worktree()
	require.NoError(t, err)

	// Write files
	for relPath, content := range files {
		fullPath := filepath.Join(repoDir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
		_, err = wt.Add(relPath)
		require.NoError(t, err)
	}

	// Write symlinks
	for linkName, target := range symlinks {
		fullPath := filepath.Join(repoDir, linkName)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.Symlink(target, fullPath))
		_, err = wt.Add(linkName)
		require.NoError(t, err)
	}

	// Commit on main/master
	sig := &object.Signature{
		Name:  "Drassi Tester",
		Email: "tester@drassi.run",
		When:  time.Now(),
	}
	mainHash, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	require.NoError(t, err)

	// Create a tag
	tagName := "v1.0.0"
	_, err = gitRepo.CreateTag(tagName, mainHash, &git.CreateTagOptions{
		Tagger:  sig,
		Message: "Release v1.0.0",
	})
	require.NoError(t, err)

	// Create and checkout a feature branch
	branchName := "feature-test"
	headRef, err := gitRepo.Head()
	require.NoError(t, err)

	err = wt.Checkout(&git.CheckoutOptions{
		Hash:   headRef.Hash(),
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})
	require.NoError(t, err)

	// Add a file on the feature branch
	featureFilePath := filepath.Join(repoDir, "feature.txt")
	require.NoError(t, os.WriteFile(featureFilePath, []byte("feature branch content"), 0o644))
	_, err = wt.Add("feature.txt")
	require.NoError(t, err)

	featHash, err := wt.Commit("Feature commit", &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	require.NoError(t, err)

	// Checkout back to main
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: headRef.Name(),
	})
	require.NoError(t, err)

	return &testGitRepoInfo{
		RepoDir:        repoDir,
		MainCommitHash: mainHash.String(),
		FeatureCommit:  featHash.String(),
		TagName:        tagName,
		BranchName:     branchName,
	}
}

func assertTarEntries(t *testing.T, r io.Reader, expectedEntries, expectedSymlinks map[string]string) {
	t.Helper()
	entries := make(map[string]string)
	symlinks := make(map[string]string)
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		if hdr.Typeflag == tar.TypeSymlink {
			symlinks[hdr.Name] = hdr.Linkname
		} else if hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			entries[hdr.Name] = string(content)
		}
	}

	if expectedEntries != nil {
		assert.Equal(t, expectedEntries, entries)
	} else {
		assert.Empty(t, entries)
	}

	if expectedSymlinks != nil {
		assert.Equal(t, expectedSymlinks, symlinks)
	} else {
		assert.Empty(t, symlinks)
	}
}

func makeLocalRepoRef(repoDir, ref string) *RepoReference {
	// Trim leading slash to form a valid name
	trimmed := strings.TrimPrefix(repoDir, "/")
	return &RepoReference{
		Scheme:    "git",
		Transport: "file",
		Endpoint:  "localhost",
		Name:      trimmed,
		Ref:       ref,
	}
}

// --- Tests ---

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tempDir := t.TempDir()
		storeDir := filepath.Join(tempDir, "store")
		mgr, err := New(storeDir)
		require.NoError(t, err)
		require.NotNil(t, mgr)

		info, err := os.Stat(storeDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("invalid dir error", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "existing-file")
		require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o644))

		// Attempting to create manager where root is a child of a file
		invalidPath := filepath.Join(filePath, "child")
		mgr, err := New(invalidPath)
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})
}

func TestFetch(t *testing.T) {
	files := map[string]string{
		"action.yml": "name: test action",
		"index.js":   "console.log('hello')",
	}
	symlinks := map[string]string{
		"link.js": "index.js",
	}
	repoInfo := initTestGitRepo(t, files, symlinks)

	t.Run("fetch default branch (HEAD)", func(t *testing.T) {
		mgr, _ := newTestManager(t)
		ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

		rev, err := mgr.Fetch(context.Background(), ref, "")
		require.NoError(t, err)
		assert.Equal(t, repoInfo.MainCommitHash, rev)
	})

	t.Run("fetch feature branch", func(t *testing.T) {
		mgr, _ := newTestManager(t)
		ref := makeLocalRepoRef(repoInfo.RepoDir, repoInfo.BranchName)

		rev, err := mgr.Fetch(context.Background(), ref, "")
		require.NoError(t, err)
		assert.Equal(t, repoInfo.FeatureCommit, rev)
	})

	t.Run("fetch tag", func(t *testing.T) {
		mgr, _ := newTestManager(t)
		ref := makeLocalRepoRef(repoInfo.RepoDir, repoInfo.TagName)

		rev, err := mgr.Fetch(context.Background(), ref, "")
		require.NoError(t, err)
		assert.Equal(t, repoInfo.MainCommitHash, rev)
	})

	t.Run("fetch non-existent ref returns error", func(t *testing.T) {
		mgr, _ := newTestManager(t)
		ref := makeLocalRepoRef(repoInfo.RepoDir, "nonexistent-branch-404")

		rev, err := mgr.Fetch(context.Background(), ref, "")
		assert.Error(t, err)
		assert.Empty(t, rev)
	})

	t.Run("fetch with auth token propagates basic auth", func(t *testing.T) {
		var receivedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		mgr, _ := newTestManager(t)
		ref := &RepoReference{
			Transport: "http",
			Endpoint:  strings.TrimPrefix(server.URL, "http://"),
			Name:      "test/repo",
			Ref:       "HEAD",
		}

		const testToken = "secret-token-123"
		_, err := mgr.Fetch(context.Background(), ref, testToken)
		assert.Error(t, err) // server returns 401
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("token:"+testToken))
		assert.Equal(t, expectedAuth, receivedAuth)
	})

	t.Run("fetch non-existent repository returns error", func(t *testing.T) {
		mgr, _ := newTestManager(t)
		ref := makeLocalRepoRef(filepath.Join(t.TempDir(), "does-not-exist"), "HEAD")

		rev, err := mgr.Fetch(context.Background(), ref, "")
		assert.Error(t, err)
		assert.Empty(t, rev)
	})
}

func TestRead(t *testing.T) {
	files := map[string]string{
		"action.yml":         "name: test-action\ndescription: test",
		"src/index.js":       "console.log('test');",
		"nested/dir/file.md": "# Docs",
	}
	symlinks := map[string]string{
		"src/alias.js": "index.js",
	}
	repoInfo := initTestGitRepo(t, files, symlinks)

	mgr, _ := newTestManager(t)
	ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")
	rev, err := mgr.Fetch(context.Background(), ref, "")
	require.NoError(t, err)

	t.Run("read full archive with directory prefix", func(t *testing.T) {
		prefix := "actions/my-action@v1"
		rc, err := mgr.Read(context.Background(), ref, rev, prefix)
		require.NoError(t, err)
		defer rc.Close()

		expectedEntries := map[string]string{
			"actions/my-action@v1/action.yml":         files["action.yml"],
			"actions/my-action@v1/src/index.js":       files["src/index.js"],
			"actions/my-action@v1/nested/dir/file.md": files["nested/dir/file.md"],
		}
		expectedSymlinks := map[string]string{
			"actions/my-action@v1/src/alias.js": "index.js",
		}

		assertTarEntries(t, rc, expectedEntries, expectedSymlinks)
	})

	t.Run("read full archive with empty prefix", func(t *testing.T) {
		rc, err := mgr.Read(context.Background(), ref, rev, "")
		require.NoError(t, err)
		defer rc.Close()

		assertTarEntries(t, rc, files, symlinks)
	})

	t.Run("read non-existent revision error", func(t *testing.T) {
		nonExistentRev := strings.Repeat("0", 40)
		rc, err := mgr.Read(context.Background(), ref, nonExistentRev, "")
		assert.Error(t, err)
		assert.Nil(t, rc)
	})

	t.Run("read non-existent repo error", func(t *testing.T) {
		unknownRef := &RepoReference{
			Endpoint: "unknown.com",
			Name:     "unknown/repo",
			Ref:      "main",
		}
		rc, err := mgr.Read(context.Background(), unknownRef, rev, "")
		assert.Error(t, err)
		assert.Nil(t, rc)
	})

	t.Run("context cancellation during read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		rc, err := mgr.Read(ctx, ref, rev, "")
		require.NoError(t, err)
		defer rc.Close()

		_, readErr := io.ReadAll(rc)
		assert.Error(t, readErr)
		assert.True(t, errors.Is(readErr, context.Canceled) || errors.Is(readErr, io.ErrClosedPipe))
	})
}

func TestFile(t *testing.T) {
	files := map[string]string{
		"action.yml":           "name: my-action",
		"src/index.js":         "console.log('ok')",
		"nested/sub/data.json": `{"key": "value"}`,
	}
	repoInfo := initTestGitRepo(t, files, nil)

	mgr, _ := newTestManager(t)
	ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")
	rev, err := mgr.Fetch(context.Background(), ref, "")
	require.NoError(t, err)

	t.Run("read existing regular file", func(t *testing.T) {
		rc, err := mgr.File(context.Background(), ref, rev, "action.yml")
		require.NoError(t, err)
		defer rc.Close()

		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, files["action.yml"], string(content))
	})

	t.Run("read nested file with leading slash and relative dots", func(t *testing.T) {
		testPaths := []string{
			"nested/sub/data.json",
			"/nested/sub/data.json",
			"./nested/sub/data.json",
			"nested/../nested/sub/data.json",
		}
		for _, p := range testPaths {
			rc, err := mgr.File(context.Background(), ref, rev, p)
			require.NoError(t, err, "path: %s", p)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			assert.Equal(t, files["nested/sub/data.json"], string(content))
			_ = rc.Close()
		}
	})

	t.Run("read non-existent file error", func(t *testing.T) {
		rc, err := mgr.File(context.Background(), ref, rev, "does-not-exist.txt")
		assert.Error(t, err)
		assert.Nil(t, rc)
		assert.True(t, errors.Is(err, object.ErrFileNotFound))
	})

	t.Run("read directory as file error", func(t *testing.T) {
		rc, err := mgr.File(context.Background(), ref, rev, "src")
		assert.Error(t, err)
		assert.Nil(t, rc)
		assert.Contains(t, err.Error(), "not a (regular) file")
	})

	t.Run("read file with invalid revision error", func(t *testing.T) {
		invalidRev := strings.Repeat("a", 40)
		rc, err := mgr.File(context.Background(), ref, invalidRev, "action.yml")
		assert.Error(t, err)
		assert.Nil(t, rc)
	})

	t.Run("read file from non-existent repo error", func(t *testing.T) {
		unknownRef := &RepoReference{
			Endpoint: "unknown.com",
			Name:     "unknown/repo",
			Ref:      "main",
		}
		rc, err := mgr.File(context.Background(), unknownRef, rev, "action.yml")
		assert.Error(t, err)
		assert.Nil(t, rc)
	})
}

func TestPersistence(t *testing.T) {
	files := map[string]string{
		"action.yml": "name: persistent-action",
	}
	repoInfo := initTestGitRepo(t, files, nil)

	rootDir := t.TempDir()

	// Manager 1: Fetches the repo
	mgr1, err := New(rootDir)
	require.NoError(t, err)
	ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")
	rev, err := mgr1.Fetch(context.Background(), ref, "")
	require.NoError(t, err)

	// Manager 2: Created with the same rootDir without calling Fetch
	mgr2, err := New(rootDir)
	require.NoError(t, err)

	// File lookup should succeed by reading from disk
	rc, err := mgr2.File(context.Background(), ref, rev, "action.yml")
	require.NoError(t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, files["action.yml"], string(content))

	// Read archive should also succeed
	tarRc, err := mgr2.Read(context.Background(), ref, rev, "")
	require.NoError(t, err)
	defer tarRc.Close()
	assertTarEntries(t, tarRc, files, nil)
}

func TestConcurrency(t *testing.T) {
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	repoInfo := initTestGitRepo(t, files, nil)

	mgr, _ := newTestManager(t)
	ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()

			// Concurrently fetch
			rev, err := mgr.Fetch(ctx, ref, "")
			require.NoError(t, err)

			// Concurrently read file
			rc, err := mgr.File(ctx, ref, rev, "file1.txt")
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			assert.Equal(t, "content1", string(content))
			_ = rc.Close()

			// Concurrently read tar
			tarRc, err := mgr.Read(ctx, ref, rev, "")
			require.NoError(t, err)
			assertTarEntries(t, tarRc, files, nil)
			_ = tarRc.Close()
		}(i)
	}

	wg.Wait()
}

func TestClose(t *testing.T) {
	files := map[string]string{
		"action.yml": "name: close-test",
	}
	repoInfo := initTestGitRepo(t, files, nil)

	mgr, _ := newTestManager(t)
	ref := makeLocalRepoRef(repoInfo.RepoDir, "HEAD")

	// Fetch repository to ensure it is opened and stored in manager
	_, err := mgr.Fetch(context.Background(), ref, "")
	require.NoError(t, err)

	// Close the manager
	err = mgr.Close()
	require.NoError(t, err)

	// Calling Close again should also succeed without error
	err = mgr.Close()
	require.NoError(t, err)
}
