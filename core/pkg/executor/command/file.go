package command

import (
	"context"
	"io"
	"os"
	"path"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/string"
	"drassi.run/core/util/tar"
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
	suffix = xstring.EnsurePrefix(suffix, "_")

	fileEntries := make(map[string]string, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		name := path.Join("file_commands", cmd+suffix)
		fileEntries[name] = ""
	}

	if r, err := xtar.ContentReader(fileEntries); err != nil {
		return err
	} else {
		return mgr.sandbox.CopyIn(ctx, r, mgr.sandbox.Layout().Temp)
	}
}

func (mgr *fileManager) Process(ctx context.Context, suffix string) error {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	suffix = xstring.EnsurePrefix(suffix, "_")
	dir := mgr.dir()

	g, ctx := errgroup.WithContext(ctx)
	for cmd, h := range mgr.registeredCommands {
		handler := h
		filePath := path.Join(dir, cmd+suffix)

		g.Go(func() error {
			r, err := mgr.sandbox.CopyOut(ctx, filePath)
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
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	suffix = xstring.EnsurePrefix(suffix, "_")
	dir := mgr.dir()

	env := make(map[string]string, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		env[cmd] = path.Join(dir, cmd+suffix)
	}
	return env
}

func (mgr *fileManager) dir() string {
	layout := mgr.sandbox.Layout()
	return path.Join(layout.Temp, "file_commands")
}

func FileRun(h *FileHandler, r io.Reader) error {
	return h.run(r)
}
