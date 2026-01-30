//go:build !windows

package ipc

import (
	"net"
	"time"
)

// listen creates a platform-appropriate listener for IPC.
// On Unix systems, this uses Unix domain sockets.
func listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}

// dial connects to the IPC server.
// On Unix systems, this uses Unix domain sockets.
func dial(path string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", path, timeout)
}
