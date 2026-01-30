package ipc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"runtime"
)

// Handler processes IPC requests from CLI clients.
// The daemon implements this interface.
type Handler interface {
	HandleStatus() StatusData
	HandleReload() (ReloadResult, error)
	HandleEnable()
	HandleDisable()
}

// Daemon is the RPC service exposed to CLI clients.
// Method names become "Daemon.Status", "Daemon.Reload", etc.
type Daemon struct {
	handler Handler
}

// Status returns the current daemon status.
func (d *Daemon) Status(_ *Empty, reply *StatusData) error {
	*reply = d.handler.HandleStatus()
	return nil
}

// Reload triggers a configuration reload.
func (d *Daemon) Reload(_ *Empty, reply *ReloadResult) error {
	result, err := d.handler.HandleReload()
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

// Enable resumes event processing.
func (d *Daemon) Enable(_ *Empty, _ *Empty) error {
	d.handler.HandleEnable()
	return nil
}

// Disable pauses event processing.
func (d *Daemon) Disable(_ *Empty, _ *Empty) error {
	d.handler.HandleDisable()
	return nil
}

// Server accepts IPC connections and serves RPC requests.
type Server struct {
	listener  net.Listener
	rpcServer *rpc.Server
}

// NewServer creates an IPC server bound to the platform-appropriate socket.
func NewServer(handler Handler) (*Server, error) {
	sockPath, err := SocketPath()
	if err != nil {
		return nil, err
	}

	// Create parent directory and remove stale socket file (Unix only)
	// Windows named pipes live in a kernel namespace, not the filesystem
	if runtime.GOOS != "windows" {
		if err := os.MkdirAll(filepath.Dir(sockPath), 0755); err != nil {
			return nil, err
		}
		os.Remove(sockPath)
	}

	listener, err := listen(sockPath)
	if err != nil {
		return nil, err
	}

	// Create RPC server and register the Daemon service
	rpcServer := rpc.NewServer()
	if err := rpcServer.RegisterName("Daemon", &Daemon{handler: handler}); err != nil {
		listener.Close()
		return nil, err
	}

	return &Server{
		listener:  listener,
		rpcServer: rpcServer,
	}, nil
}

// Serve accepts connections until the context is cancelled.
// Requests are processed serially (one at a time).
func (s *Server) Serve(ctx context.Context) error {
	// Close listener and clean up socket when context is done
	go func() {
		<-ctx.Done()
		s.listener.Close()
		// Remove Unix socket file
		if runtime.GOOS != "windows" {
			if sockPath, err := SocketPath(); err == nil {
				os.Remove(sockPath)
			}
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if we're shutting down
			if ctx.Err() != nil {
				break
			}
			// Log and continue on transient errors
			if !errors.Is(err, net.ErrClosed) {
				slog.Warn("ipc accept error", "error", err)
			}
			continue
		}

		// Handle connection serially (blocks until client disconnects)
		s.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
	}

	return nil
}
