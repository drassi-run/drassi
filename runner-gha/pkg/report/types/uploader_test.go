/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestAzureBlobUploaderSuite(t *testing.T) {
	suite.Run(t, new(AzureBlobUploaderTestSuite))
}

type AzureBlobUploaderTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	server   *httptest.Server
	uploader *azureBlobUploader

	onCreate http.HandlerFunc
}

func (s *AzureBlobUploaderTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("x-ms-blob-type") != "BlockBlob" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if s.onCreate != nil {
			s.onCreate(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	s.uploader = &azureBlobUploader{
		getUrl: func(ctx context.Context) (string, error) {
			return s.server.URL, nil
		},
	}

	s.ctrl = gomock.NewController(s.T())
}

func (s *AzureBlobUploaderTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AzureBlobUploaderTestSuite) TestUpload_Success() {
	r := strings.NewReader(`hello world`)

	var created atomic.Bool
	s.onCreate = func(w http.ResponseWriter, r *http.Request) {
		content, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		s.EqualValues("hello world", content)
		created.Store(true)
		w.WriteHeader(http.StatusCreated)
	}

	err := s.uploader.Upload(s.T().Context(), r, nil)
	s.Require().NoError(err)
	s.True(created.Load())
}

type mockUploader struct {
	Uploader
	mock.Mock
}

func (u *mockUploader) Upload(ctx context.Context, r io.Reader, stat *Stat) error {
	args := u.Called(ctx)
	return args.Error(0)
}

func TestStorageAwareUploader(t *testing.T) {
	t.Run("azure_blob", func(t *testing.T) {
		s := NewStorageAwareUploader(func(context.Context) (SignedUrlResponse, error) {
			r := &mockSignedUrl{StorageType: StorageAzureBlob}
			return r, nil
		}).(*storageAwareUploader)

		u, err := s.underlay(t.Context())
		require.NoError(t, err)
		assert.IsType(t, new(azureBlobUploader), u)
	})

	t.Run("invalid", func(t *testing.T) {
		s := NewStorageAwareUploader(func(context.Context) (SignedUrlResponse, error) {
			r := &mockSignedUrl{StorageType: "UNSUPPORTED"}
			return r, nil
		}).(*storageAwareUploader)

		_, err := s.underlay(t.Context())
		require.ErrorContains(t, err, "unsupported storage type")
	})

	t.Run("getUrl-error", func(t *testing.T) {
		s := NewStorageAwareUploader(func(context.Context) (SignedUrlResponse, error) {
			return nil, errors.New("getUrl error")
		}).(*storageAwareUploader)

		_, err := s.underlay(t.Context())
		require.ErrorContains(t, err, "getUrl error")
	})

	t.Run("upload-success", func(t *testing.T) {
		ctx := t.Context()

		up := new(mockUploader)
		up.On("Upload", ctx, mock.Anything, mock.Anything).Return(nil)
		u := &storageAwareUploader{Uploader: up}

		err := u.Upload(ctx, nil, nil)
		require.NoError(t, err)
	})

	t.Run("upload-error", func(t *testing.T) {
		ctx := t.Context()

		up := new(mockUploader)
		up.On("Upload", ctx, mock.Anything, mock.Anything).Return(errors.New("upload error"))
		u := &storageAwareUploader{Uploader: up}

		err := u.Upload(ctx, nil, nil)
		require.ErrorContains(t, err, "upload error")
	})
}
