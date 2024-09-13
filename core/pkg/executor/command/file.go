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
	Initialize(ctx context.Context, sandbox sandboxer.Sandbox) (map[string]string, error)
	Process(ctx context.Context, sandbox sandboxer.Sandbox) error
}

func NewFileManager(suffix string) FileManager {
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}

	return &fileManager{
		registeredCommands: make(map[string]*FileHandler),
		suffix:             suffix,
	}
}

type fileManager struct {
	registeredCommands map[string]*FileHandler

	suffix string
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

func (mgr *fileManager) Initialize(ctx context.Context, sandbox sandboxer.Sandbox) (map[string]string, error) {
	if len(mgr.registeredCommands) == 0 {
		return nil, nil
	}

	dir := filepath.Join(sandbox.GetTempDir(), "file_commands")
	env := make(map[string]string, len(mgr.registeredCommands))
	fileEntries := make([]*utilreader.FileEntry, 0, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		name := cmd + mgr.suffix
		env[cmd] = filepath.Join(dir, name)
		fileEntries = append(fileEntries, &utilreader.FileEntry{
			Name: name,
			Mode: 0o666,
		})
	}

	if r, err := utilreader.FromFileEntries(ctx, fileEntries...); err != nil {
		return nil, err
	} else if err = sandbox.CopyIn(ctx, r, dir); err != nil {
		return nil, err
	}
	return env, nil
}

func (mgr *fileManager) Process(ctx context.Context, sandbox sandboxer.Sandbox) error {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}

	dir := filepath.Join(sandbox.GetTempDir(), "file_commands")

	g, ctx := errgroup.WithContext(ctx)
	for cmd, h := range mgr.registeredCommands {
		handler := h
		path := filepath.Join(dir, cmd+mgr.suffix)

		g.Go(func() error {
			r, err := sandbox.CopyOut(ctx, path)
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
