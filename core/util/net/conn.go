package xnet

import (
	"errors"
	"io"
	"net"
	"time"
)

const (
	localAddr  dummyAddr = "dummy-local"
	remoteAddr dummyAddr = "dummy-remote"
)

type dummyAddr string

func (d dummyAddr) Network() string { return "dummy" }
func (d dummyAddr) String() string  { return string(d) }

func NewStdioConn(stdin io.WriteCloser, stdout io.ReadCloser) net.Conn {
	conn := &stdioConn{
		stdin:  stdin,
		stdout: stdout,
	}

	return conn
}

// - [github.com/docker/cli/cli/connhelper/commandconn.New]
// - [github.com/containers/podman/v5/pkg/bindings.NewConnectionWithIdentity]
// - [golang.org/x/crypto/ssh.(*Client).Dial]
type stdioConn struct {
	stdin  io.WriteCloser // conn writing to stdin
	stdout io.ReadCloser  // conn reading from stdout
}

func (s *stdioConn) Read(b []byte) (n int, err error) {
	return s.stdout.Read(b)
}

func (s *stdioConn) Write(b []byte) (n int, err error) {
	return s.stdin.Write(b)
}

func (s *stdioConn) Close() error {
	errs := make([]error, 2)
	errs[0] = s.stdin.Close()
	errs[1] = s.stdout.Close()

	return errors.Join(errs...)
}

func (s *stdioConn) LocalAddr() net.Addr {
	return localAddr
}

func (s *stdioConn) RemoteAddr() net.Addr {
	return remoteAddr
}

func (s *stdioConn) SetDeadline(t time.Time) error {
	return errors.New("stdioConn: deadline not supported")
}

func (s *stdioConn) SetReadDeadline(t time.Time) error {
	return errors.New("stdioConn: deadline not supported")
}

func (s *stdioConn) SetWriteDeadline(t time.Time) error {
	return errors.New("stdioConn: deadline not supported")
}
