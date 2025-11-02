package report

import (
	"context"
	"fmt"
	"io"

	"drassi.run/gha-runner/pkg/log"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/appendblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

// Conveyor used to continuous upload files to cloud storage by chunks
type Conveyor interface {
	io.Closer
	Update(u *log.Update)
	Run(ctx context.Context) (*Stat, error)
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
