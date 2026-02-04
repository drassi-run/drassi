package service

import (
	"context"
	"fmt"
	"io"

	"drassi.run/gha-runner/pkg/reporter/log"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/appendblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type Result struct {
	lines int64
	size  int64
}

// Conveyor used to continuous upload files to cloud storage by chunks
type Conveyor interface {
	io.Closer
	Update(u *log.Update)
	Run(ctx context.Context) (*Result, error)
}

type storageManagerConveyor struct {
	Conveyor
	getUrl func(context.Context) (SignedUrlResponse, error)
}

func NewStorageMangerConveyor(f func(context.Context) (SignedUrlResponse, error)) Conveyor {
	return &storageManagerConveyor{getUrl: f}
}

func (s *storageManagerConveyor) Run(ctx context.Context) (*Result, error) {
	if c, err := s.underlay(ctx); err != nil {
		return nil, err
	} else {
		return c.Run(ctx)
	}
}

func (s *storageManagerConveyor) underlay(ctx context.Context) (Conveyor, error) {
	if s.Conveyor == nil {
		r, err := s.getUrl(ctx)
		if err != nil {
			return nil, err
		}

		softLimit := int64(10) * 1024 * 1024 // 10MBi
		if i, ok := r.(interface{ GetSoftSizeLimit() int64 }); ok {
			softLimit = i.GetSoftSizeLimit()
		}
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

func (s *storageManagerConveyor) getUrlString(ctx context.Context) (string, error) {
	if r, err := s.getUrl(ctx); err != nil {
		return "", err
	} else {
		return r.GetUrl(), nil
	}
}

type azureBlobConveyor struct {
	chunker *log.Chunker
	getUrl  func(context.Context) (string, error)
}

func (c *azureBlobConveyor) Update(u *log.Update) {
	c.chunker.Update(u)
}

func (c *azureBlobConveyor) Close() error {
	return c.chunker.Close()
}

func (c *azureBlobConveyor) Run(ctx context.Context) (*Result, error) {
	if err := c.create(ctx); err != nil {
		return nil, fmt.Errorf("failed to create azure blob: %w", err)
	}

	r := new(Result)
	for chunk := range c.chunker.Channel() {
		if err := c.upload(ctx, chunk); err != nil {
			return nil, fmt.Errorf("error while upload log chunk to azure blob: %w", err)
		}
		r.lines += chunk.Lines()
		r.size += chunk.Size()
	}

	if err := c.seal(ctx); err != nil {
		return nil, fmt.Errorf("error while sealing azure blob: %w", err)
	}
	return r, nil
}

func (c *azureBlobConveyor) create(ctx context.Context) error {
	var textPlain = "text/plain"
	o := &appendblob.CreateOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &textPlain,
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
