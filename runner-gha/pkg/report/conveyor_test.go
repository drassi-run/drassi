package report

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drassi.run/gha-runner/mock/log"
	"drassi.run/gha-runner/pkg/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestAzureBlobConveyorTestSuite(t *testing.T) {
	suite.Run(t, new(AzureBlobConveyorTestSuite))
}

type AzureBlobConveyorTestSuite struct {
	suite.Suite
	server *httptest.Server

	ctrl     *gomock.Controller
	chunker  *mock_log.MockChunker
	conveyor *azureBlobConveyor

	onCreate      http.HandlerFunc
	onAppendBlock http.HandlerFunc
	onSeal        http.HandlerFunc
}

func (s *AzureBlobConveyorTestSuite) SetupTest() {
	s.onCreate = nil
	s.onAppendBlock = nil
	s.onSeal = nil
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch comp := r.URL.Query().Get("comp"); comp {
		case "": // Create
			if r.Header.Get("x-ms-blob-type") != "AppendBlob" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if s.onCreate != nil {
				s.onCreate(w, r)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case "appendblock":
			if s.onAppendBlock != nil {
				s.onAppendBlock(w, r)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case "seal":
			if s.onSeal != nil {
				s.onSeal(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))

	s.ctrl = gomock.NewController(s.T())
	s.chunker = mock_log.NewMockChunker(s.ctrl)
	s.conveyor = &azureBlobConveyor{
		chunker: s.chunker,
		getUrl: func(ctx context.Context) (string, error) {
			return s.server.URL + "/myblob", nil
		},
	}
}

func (s *AzureBlobConveyorTestSuite) TearDownTest() {
	s.ctrl.Finish()
	s.server.Close()
}

func (s *AzureBlobConveyorTestSuite) TestRun_Success() {
	ctx := s.T().Context()

	ch := make(chan log.Chunk, 100)
	ch <- s.mockChunk("hello")
	ch <- s.mockChunk("world")
	close(ch)

	s.chunker.EXPECT().Channel().Return(ch)

	stat, err := s.conveyor.Run(ctx)
	s.NoError(err)
	s.NotNil(stat)
	s.EqualValues(12, stat.Size)
	s.Equal(2, stat.Lines)
}

func (s *AzureBlobConveyorTestSuite) TestRun_CreateError() {
	ctx := s.T().Context()
	s.onCreate = azureBlobError(409, "BlobAlreadyExists", "Blob already exists.")

	stat, err := s.conveyor.Run(ctx)
	s.ErrorContains(err, "failed to create azure blob")
	s.Nil(stat)
}

func (s *AzureBlobConveyorTestSuite) TestRun_UploadError() {
	ctx := s.T().Context()

	ch := make(chan log.Chunk, 1)
	ch <- s.mockChunk("hello")
	close(ch)

	s.chunker.EXPECT().Channel().Return(ch)
	s.onAppendBlock = azureBlobError(404, "BlobNotFound", "The specified blob does not exist.")

	stat, err := s.conveyor.Run(ctx)
	s.Error(err)
	s.Contains(err.Error(), "error while upload log chunk to azure blob")
	s.Nil(stat)
}

func (s *AzureBlobConveyorTestSuite) TestRun_SealError() {
	ctx := s.T().Context()
	ch := make(chan log.Chunk)
	close(ch)

	s.chunker.EXPECT().Channel().Return(ch)
	s.onSeal = azureBlobError(400, "SealFailed", "Failed to seal blob.")

	stat, err := s.conveyor.Run(ctx)
	s.Error(err)
	s.Contains(err.Error(), "error while sealing azure blob")
	s.Nil(stat)
}

func (s *AzureBlobConveyorTestSuite) TestUpdate() {
	u := new(log.Update)
	// Expect Update is called with param: u
	s.chunker.EXPECT().Update(u)

	s.conveyor.Update(u)
}

func (s *AzureBlobConveyorTestSuite) TestClose() {
	s.chunker.EXPECT().Close().Return(nil)
	err := s.conveyor.Close()
	s.NoError(err)
}

func (s *AzureBlobConveyorTestSuite) mockChunk(content string) log.Chunk {
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	r := streaming.NopCloser(strings.NewReader(content))
	size := int64(len(content))
	lines := strings.Count(content, "\n")

	chunk := mock_log.NewMockChunk(s.ctrl)
	chunk.EXPECT().Empty().Return(false).AnyTimes()
	chunk.EXPECT().Size().Return(size).AnyTimes()
	chunk.EXPECT().Lines().Return(lines).AnyTimes()
	chunk.EXPECT().Reader().Return(r, nil).AnyTimes()

	return chunk
}

// Should NOT using httpCode=500 to avoid client retry
// https://learn.microsoft.com/en-us/rest/api/storageservices/blob-service-error-codes
func azureBlobError(httpCode int, errorCode, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(httpCode)

		fmt.Fprintf(w, "<Error>\n\t<Code>%s</Code>\n\t<Message>%s</Message>\n</Error>", errorCode, message)
	}
}
