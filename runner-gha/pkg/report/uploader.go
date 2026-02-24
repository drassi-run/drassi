package report

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
)

const StorageAzureBlob = "BLOB_STORAGE_TYPE_AZURE"

// Uploader used to one-shot upload a file to cloud storage
type Uploader interface {
	Upload(ctx context.Context, r io.Reader, stat *Stat) error
}

type storageAwareUploader struct {
	Uploader
	getUrl func(context.Context) (SignedUrlResponse, error)
}

func NewStorageAwareUploader(f func(context.Context) (SignedUrlResponse, error)) Uploader {
	return &storageAwareUploader{getUrl: f}
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

type azureBlobUploader struct {
	getUrl func(context.Context) (string, error)
}

func (u *azureBlobUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	if client, err := u.getClient(ctx); err != nil {
		return err
	} else {
		_, err = client.UploadStream(ctx, r, nil)
		return err
	}
}

func (u *azureBlobUploader) getClient(ctx context.Context) (*blockblob.Client, error) {
	if url, err := u.getUrl(ctx); err != nil {
		return nil, err
	} else {
		return blockblob.NewClientWithNoCredential(url, nil)
	}
}
