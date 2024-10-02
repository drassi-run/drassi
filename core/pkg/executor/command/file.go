package command

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	utilreader "drassi.run/core/pkg/util/reader"
	"golang.org/x/sync/errgroup"
)

type FileHandler struct {
	env string
	run func(r io.Reader) error
}

func NewFileHandler(env string, run func(r io.Reader) error) *FileHandler {
	return &FileHandler{
		env: env,
		run: run,
	}
}

type FileManager interface {
	Register(handler *FileHandler) error
	Initialize(ctx context.Context, suffix string) error
	Process(ctx context.Context, suffix string) error
	Env(suffix string) map[string]string
}

func NewFileManager(sandbox sandboxer.Sandbox) FileManager {
	return &fileManager{
		registeredCommands: make(map[string]*FileHandler),
		sandbox:            sandbox,
	}
}

type fileManager struct {
	sandbox            sandboxer.Sandbox
	registeredCommands map[string]*FileHandler
}

func (mgr *fileManager) Register(handler *FileHandler) error {
	env := handler.env
	if handler.run == nil {
		delete(mgr.registeredCommands, env)
	} else {
		mgr.registeredCommands[env] = handler
	}
	return nil
}

func (mgr *fileManager) Initialize(ctx context.Context, suffix string) error {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}

	dir := filepath.Join(mgr.sandbox.GetTempDir(), "file_commands")
	fileEntries := make([]*utilreader.FileEntry, 0, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		fileEntries = append(fileEntries, &utilreader.FileEntry{
			Name: cmd + suffix,
			Mode: 0o666,
		})
	}

	if r, err := utilreader.FromFileEntries(ctx, fileEntries...); err != nil {
		return err
	} else {
		return mgr.sandbox.CopyIn(ctx, r, dir)
	}
}

func (mgr *fileManager) Process(ctx context.Context, suffix string) error {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}

	dir := filepath.Join(mgr.sandbox.GetTempDir(), "file_commands")

	g, ctx := errgroup.WithContext(ctx)
	for cmd, h := range mgr.registeredCommands {
		handler := h
		path := filepath.Join(dir, cmd+suffix)

		g.Go(func() error {
			r, err := mgr.sandbox.CopyOut(ctx, path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			defer r.Close()
			return handler.run(r)
		})
	}
	return g.Wait()
}

func (mgr *fileManager) Env(suffix string) map[string]string {
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}

	dir := filepath.Join(mgr.sandbox.GetTempDir(), "file_commands")
	env := make(map[string]string, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		env[cmd] = filepath.Join(dir, cmd+suffix)
	}
	return env
}

func FileRun(h *FileHandler, r io.Reader) error {
	return h.run(r)
}
