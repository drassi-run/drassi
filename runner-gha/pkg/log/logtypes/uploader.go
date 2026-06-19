/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/chainguard-dev/clog"
)

// Uploader used to one-shot upload a file to cloud storage
type Uploader interface {
	Upload(ctx context.Context, r io.Reader, stat *Stat) error
}

func NewStorageAwareUploader(f func(context.Context) (SignedUrlResponse, error)) Uploader {
	return &storageAwareUploader{getUrl: f}
}

// storageAwareUploader implement Uploader that selects the appropriate storage backend
// (e.g., S3, Azure Blob, or GCS) based on response of getUrl.
type storageAwareUploader struct {
	Uploader
	getUrl func(context.Context) (SignedUrlResponse, error)
}

func (s *storageAwareUploader) Upload(ctx context.Context, r io.Reader, stat *Stat) error {
	if u, err := s.underlay(ctx); err != nil {
		return err
	} else {
		return u.Upload(ctx, r, stat)
	}
}

func (s *storageAwareUploader) underlay(ctx context.Context) (Uploader, error) {
	if s.Uploader == nil {
		r, err := s.getUrl(ctx)
		if err != nil {
			return nil, err
		}

		switch typ := r.GetStorageType(); typ {
		case StorageAzureBlob:
			s.Uploader = &azureBlobUploader{s.getUrlString}
		default:
			return nil, fmt.Errorf("unsupported storage type %s", typ)
		}
	}

	return s.Uploader, nil
}

func (s *storageAwareUploader) getUrlString(ctx context.Context) (string, error) {
	if r, err := s.getUrl(ctx); err != nil {
		return "", err
	} else {
		return r.GetUrl(), nil
	}
}

// azureBlobUploader is Uploader implementation for Azure Blob storage backend.
// It uploads into [BlockBlob].
//
// [BlockBlob]: https://learn.microsoft.com/en-us/rest/api/storageservices/operations-on-block-blobs
type azureBlobUploader struct {
	getUrl func(context.Context) (string, error)
}

func (u *azureBlobUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	if client, err := u.getClient(ctx); err != nil {
		return err
	} else {
		clog.DebugContext(ctx, "AzureBlob - upload block blob")
		if _, err = client.UploadStream(ctx, r, nil); err != nil {
			return fmt.Errorf("upload to azure blob: %w", err)
		}
		return nil
	}
}

func (u *azureBlobUploader) getClient(ctx context.Context) (*blockblob.Client, error) {
	if url, err := u.getUrl(ctx); err != nil {
		return nil, err
	} else {
		return blockblob.NewClientWithNoCredential(url, nil)
	}
}

type FuncUploader func(ctx context.Context, r io.Reader, stat *Stat) error

func (f FuncUploader) Upload(ctx context.Context, r io.Reader, stat *Stat) error {
	return f(ctx, r, stat)
}
