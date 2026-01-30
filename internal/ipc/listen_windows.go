//go:build windows

package ipc

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// listen creates a platform-appropriate listener for IPC.
// On Windows, this uses named pipes.
func listen(path string) (net.Listener, error) {
	return winio.ListenPipe(path, nil)
}

// dial connects to the IPC server.
// On Windows, this uses named pipes.
func dial(path string, timeout time.Duration) (net.Conn, error) {
	return winio.DialPipe(path, &timeout)
}
