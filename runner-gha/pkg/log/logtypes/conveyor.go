/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import (
	"context"
	"fmt"

	"drassi.run/gha-runner/pkg/log"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/appendblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

// Conveyor used to continuous upload files to cloud storage by chunks
type Conveyor interface {
	Update(u *log.Update)
	Run(ctx context.Context) (*Stat, error)
	Close() error
}

func NewStorageAwareConveyor(f func(context.Context) (SignedUrlResponse, error)) Conveyor {
	return &storageAwareConveyor{getUrl: f}
}

// storageAwareConveyor implement Conveyor that selects the appropriate storage backend
// (e.g., S3, Azure Blob, or GCS) based on response of getUrl.
type storageAwareConveyor struct {
	Conveyor
	getUrl func(context.Context) (SignedUrlResponse, error)
}

func (s *storageAwareConveyor) Close() error {
	if c := s.Conveyor; c != nil {
		return c.Close()
	}
	return nil
}

func (s *storageAwareConveyor) Run(ctx context.Context) (*Stat, error) {
	if c, err := s.underlay(ctx); err != nil {
		return nil, err
	} else {
		return c.Run(ctx)
	}
}

func (s *storageAwareConveyor) underlay(ctx context.Context) (Conveyor, error) {
	if s.Conveyor == nil {
		r, err := s.getUrl(ctx)
		if err != nil {
			return nil, err
		}

		softLimit := int64(10) * 1024 * 1024 // 10MBi
		if i, ok := r.(interface{ GetSoftSizeLimit() int64 }); ok {
			softLimit = i.GetSoftSizeLimit()
		}
		//goland:noinspection GoResourceLeak
		chunker := log.NewChunker(softLimit)

		switch typ := r.GetStorageType(); typ {
		case StorageAzureBlob:
			s.Conveyor = &azureBlobConveyor{chunker, s.getUrlString}
		default:
			return nil, fmt.Errorf("unsupported storage type %s", typ)
		}
	}

	return s.Conveyor, nil
}

func (s *storageAwareConveyor) getUrlString(ctx context.Context) (string, error) {
	if r, err := s.getUrl(ctx); err != nil {
		return "", err
	} else {
		return r.GetUrl(), nil
	}
}

// azureBlobConveyor is Conveyor implementation for Azure Blob storage backend.
// It upload Chunks into [AppendBlob].
//
// [AppendBlob]: https://learn.microsoft.com/en-us/rest/api/storageservices/operations-on-append-blobs
type azureBlobConveyor struct {
	chunker log.Chunker
	getUrl  func(context.Context) (string, error)
}

func (c *azureBlobConveyor) Update(u *log.Update) {
	c.chunker.Update(u)
}

func (c *azureBlobConveyor) Close() error {
	return c.chunker.Close()
}

func (c *azureBlobConveyor) Run(ctx context.Context) (*Stat, error) {
	if err := c.create(ctx); err != nil {
		return nil, fmt.Errorf("failed to create azure blob: %w", err)
	}

	s := new(Stat)
	for chunk := range c.chunker.Channel() {
		if err := c.upload(ctx, chunk); err != nil {
			return nil, fmt.Errorf("error while upload log chunk to azure blob: %w", err)
		}
		s.Lines += chunk.Lines()
		s.Size += chunk.Size()
	}

	if err := c.seal(ctx); err != nil {
		return nil, fmt.Errorf("error while sealing azure blob: %w", err)
	}
	return s, nil
}

func (c *azureBlobConveyor) create(ctx context.Context) error {
	o := &appendblob.CreateOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: new("text/plain"),
		},
	}

	if client, err := c.getClient(ctx); err != nil {
		return err
	} else {
		_, err = client.Create(ctx, o)
		return err
	}
}

func (c *azureBlobConveyor) upload(ctx context.Context, chunk log.Chunk) error {
	if chunk.Empty() {
		return nil
	}

	if client, err := c.getClient(ctx); err != nil {
		return err
	} else if r, err := chunk.Reader(); err != nil {
		return err
	} else {
		defer r.Close() // ensure r is closed, even AppendBlock error
		_, err = client.AppendBlock(ctx, r, nil)
		return err
	}
}

func (c *azureBlobConveyor) seal(ctx context.Context) error {
	if client, err := c.getClient(ctx); err != nil {
		return err
	} else {
		_, err = client.Seal(ctx, nil)
		return err
	}
}

func (c *azureBlobConveyor) getClient(ctx context.Context) (*appendblob.Client, error) {
	if url, err := c.getUrl(ctx); err != nil {
		return nil, err
	} else {
		return appendblob.NewClientWithNoCredential(url, nil)
	}
}
