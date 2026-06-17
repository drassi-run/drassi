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
	"drassi.run/core/util/tar"
	"github.com/docker/docker/pkg/archive"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidFile    = errors.New("invalid file")
	ErrorMultipleFile = errors.New("un-expected multiple files")
)

type FileRun[R any] func(ctx context.Context, res R, r io.Reader) error

type FileHandler[R any] struct {
	env string
	run FileRun[R]
}

func (h *FileHandler[R]) Run(ctx context.Context, res R, r io.Reader) error {
	return h.run(ctx, res, r)
}

func NewFileHandler[R any](env string, run FileRun[R]) *FileHandler[R] {
	return &FileHandler[R]{
		env: env,
		run: run,
	}
}

type Filer interface {
	CommandFile(cmd string) string
}

type FileManager[R any] interface {
	Register(handler *FileHandler[R]) error
	Initialize(ctx context.Context, res R) error
	Process(ctx context.Context, res R) error
	Env(res R) map[string]string
}

func NewFileManager[R any](sandbox sandboxer.Sandbox) FileManager[R] {
	return &fileManager[R]{
		registeredCommands: make(map[string]*FileHandler[R]),
		sandbox:            sandbox,
	}
}

type fileManager[R any] struct {
	sandbox            sandboxer.Sandbox
	registeredCommands map[string]*FileHandler[R]
}

func (mgr *fileManager[R]) Register(handler *FileHandler[R]) error {
	env := handler.env
	if handler.run == nil {
		delete(mgr.registeredCommands, env)
	} else {
		mgr.registeredCommands[env] = handler
	}
	return nil
}

func (mgr *fileManager[R]) Initialize(ctx context.Context, res R) (err error) {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}

	ctx, span := xotel.StartSpan(ctx, "FileCommand.Initialize")
	defer xotel.EndSpan(span, &err)

	fileEntries := make(map[string]string, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		name := path.Join("file_commands", mgr.pathOf(res, cmd))
		fileEntries[name] = ""
	}

	if r, err := xtar.ContentReader(fileEntries); err != nil {
		return err
	} else {
		return mgr.sandbox.CopyIn(ctx, r, mgr.sandbox.Layout().Temp)
	}
}

func (mgr *fileManager[R]) Process(ctx context.Context, res R) (err error) {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	dir := mgr.dir()

	ctx, span := xotel.StartSpan(ctx, "FileCommand.Process")
	defer xotel.EndSpan(span, &err)

	g, ctx := errgroup.WithContext(ctx)
	for cmd, h := range mgr.registeredCommands {
		handler := h
		filePath := path.Join(dir, mgr.pathOf(res, cmd))

		g.Go(func() error {
			return mgr.handle(ctx, res, handler, filePath)
		})
	}
	return g.Wait()
}

func (mgr *fileManager[R]) handle(ctx context.Context, res R, handler *FileHandler[R], file string) error {
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
		if err = handler.run(ctx, res, tr); err != nil {
			return err
		}
	}

	if _, err = tr.Next(); err != io.EOF {
		return ErrorMultipleFile
	}
	return nil
}

func (mgr *fileManager[R]) Env(res R) map[string]string {
	if len(mgr.registeredCommands) == 0 {
		return nil
	}
	dir := mgr.dir()

	env := make(map[string]string, len(mgr.registeredCommands))
	for cmd := range mgr.registeredCommands {
		env[cmd] = path.Join(dir, mgr.pathOf(res, cmd))
	}
	return env
}

func (mgr *fileManager[R]) dir() string {
	layout := mgr.sandbox.Layout()
	return path.Join(layout.Temp, "file_commands")
}

func (mgr *fileManager[R]) pathOf(res any, cmd string) string {
	if f, ok := res.(Filer); ok {
		return f.CommandFile(cmd)
	}
	return cmd
}
