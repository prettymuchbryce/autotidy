package ipc

import (
	"log/slog"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

// Client connects to the daemon via JSON-RPC over a Unix socket.
type Client struct {
	rpc *rpc.Client
}

// Connect establishes a connection to the daemon.
// Returns an error if the daemon is not running.
func Connect() (*Client, error) {
	sockPath, err := SocketPath()
	if err != nil {
		return nil, err
	}
	conn, err := dial(sockPath, 2*time.Second)
	if err != nil {
		slog.Error("Failed to connect to the autotidy daemon. Are you sure it's running?", "error", err)
		return nil, err
	}
	return &Client{rpc: jsonrpc.NewClient(conn)}, nil
}

// Status queries the daemon for its current status.
func (c *Client) Status() (*StatusData, error) {
	var status StatusData
	if err := c.rpc.Call("Daemon.Status", &Empty{}, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Reload tells the daemon to reload its configuration.
func (c *Client) Reload() (*ReloadResult, error) {
	var result ReloadResult
	if err := c.rpc.Call("Daemon.Reload", &Empty{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Enable tells the daemon to resume event processing.
func (c *Client) Enable() error {
	return c.rpc.Call("Daemon.Enable", &Empty{}, &Empty{})
}

// Disable tells the daemon to pause event processing.
func (c *Client) Disable() error {
	return c.rpc.Call("Daemon.Disable", &Empty{}, &Empty{})
}

// Close closes the connection to the daemon.
func (c *Client) Close() error {
	return c.rpc.Close()
}
