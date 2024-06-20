package docker

import (
	"context"
	"io"
	"os"

	"drassi.run/core/pkg/util"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

func streamingResponse(ctx context.Context, resp dockertypes.HijackedResponse, isTerminal bool) error {
	logger := util.Logger(ctx)
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		var outWriter io.Writer = os.Stdout
		var errWriter io.Writer = os.Stderr

		var err error
		if !isTerminal || os.Getenv("NORAW") != "" {
			_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
		} else {
			_, err = io.Copy(outWriter, resp.Reader)
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		_, err := resp.Conn.Write([]byte{3}) // send ctrl + c
		if err != nil {
			logger.Warnf("Failed to send CTRL+C: %+s", err)
		}

		// we return the context canceled error to prevent other steps from executing
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			logger.Error(err)
		}

		return nil
	}
}
