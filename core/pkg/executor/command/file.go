/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/otel"
	"drassi.run/core/util/string"
	"drassi.run/core/util/tar"
	"github.com/docker/docker/pkg/archive"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidFile    = errors.New("invalid file")
	ErrorMultipleFile = errors.New("un-expected multiple files")
)

type FileHandler struct {
	env string
	run func(ctx context.Context, r io.Reader) error
}

func NewFileHandler(env string, run func(context.Context, io.Reader) error) *FileHandler {
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

func (mgr *fileManager) Initialize(ctx context.Context, suffix string) (err error) {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	suffix = xstring.EnsurePrefix(suffix, "_")

	ctx, span := xotel.StartSpan(ctx, "FileCommand.Initialize")
	defer xotel.EndSpan(span, &err)

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

func (mgr *fileManager) Process(ctx context.Context, suffix string) (err error) {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	suffix = xstring.EnsurePrefix(suffix, "_")
	dir := mgr.dir()

	ctx, span := xotel.StartSpan(ctx, "FileCommand.Process")
	defer xotel.EndSpan(span, &err)

	g, ctx := errgroup.WithContext(ctx)
	for cmd, h := range mgr.registeredCommands {
		handler := h
		filePath := path.Join(dir, cmd+suffix)

		g.Go(func() error {
			return mgr.handle(ctx, handler, filePath)
		})
	}
	return g.Wait()
}

func (mgr *fileManager) handle(ctx context.Context, handler *FileHandler, file string) error {
	r, err := mgr.sandbox.CopyOut(ctx, file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer r.Close()

	xr, err := archive.DecompressStream(r)
	if err != nil {
		return err
	}
	defer xr.Close()

	tr := tar.NewReader(xr)
	if hdr, err := tr.Next(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	} else if !xtar.IsRegular(hdr) {
		return fmt.Errorf("%w: un-expected %s file", ErrInvalidFile, xtar.FileType(hdr.Typeflag))
	} else if hdr.Size > 0 {
		if err = handler.run(ctx, tr); err != nil {
			return err
		}
	}

	if _, err = tr.Next(); err != io.EOF {
		return ErrorMultipleFile
	}
	return nil
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

func FileRun(ctx context.Context, h *FileHandler, r io.Reader) error {
	return h.run(ctx, r)
}
