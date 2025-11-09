package report

import (
	"context"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
)

// Uploader used to one-shot upload a file to cloud storage
type Uploader interface {
	Upload(ctx context.Context, r io.Reader, stat *Stat) error
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
