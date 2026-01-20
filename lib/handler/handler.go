// Package handler implements SAM command handlers per SAMv3.md specification.
// Each handler processes a specific SAM command (HELLO, SESSION, STREAM, etc.)
// and returns an appropriate response.
package handler

import (
	"context"
	"net"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
	"github.com/go-i2p/go-sam-bridge/lib/session"
)

// Handler processes a SAM command and returns a response.
// Implementations must be safe for concurrent use.
type Handler interface {
	// Handle processes the command and returns a response.
	// Returns nil response if no response should be sent (e.g., after QUIT).
	// Returns error for internal errors (connection issues, not protocol errors).
	Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error)
}

// HandlerFunc is a function adapter for Handler interface.
// Allows using functions as handlers without creating a struct.
type HandlerFunc func(ctx *Context, cmd *protocol.Command) (*protocol.Response, error)

// Handle implements Handler by calling the function.
func (f HandlerFunc) Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
	return f(ctx, cmd)
}

// Context holds state for command execution.
// Created per-command and contains connection-specific information.
type Context struct {
	// Conn is the client connection.
	Conn net.Conn

	// Session is the bound session, if any.
	// Nil until SESSION CREATE succeeds on this connection.
	Session session.Session

	// Registry provides access to the global session registry.
	Registry session.Registry

	// Version is the negotiated SAM version after HELLO.
	// Empty string before handshake completes.
	Version string

	// Authenticated indicates if the client has authenticated.
	// Always true if authentication is disabled on the bridge.
	Authenticated bool

	// HandshakeComplete indicates if HELLO has been received.
	HandshakeComplete bool

	// Ctx is the request context for cancellation and timeouts.
	Ctx context.Context
}

// NewContext creates a new handler context with the given connection.
func NewContext(conn net.Conn, registry session.Registry) *Context {
	return &Context{
		Conn:     conn,
		Registry: registry,
		Ctx:      context.Background(),
	}
}

// WithContext returns a copy of the Context with the given context.Context.
func (c *Context) WithContext(ctx context.Context) *Context {
	newCtx := *c
	newCtx.Ctx = ctx
	return &newCtx
}

// BindSession binds a session to this connection context.
func (c *Context) BindSession(s session.Session) {
	c.Session = s
}

// UnbindSession removes the session binding from this context.
func (c *Context) UnbindSession() {
	c.Session = nil
}

// RemoteAddr returns the remote address of the client connection.
// Returns empty string if connection is nil.
func (c *Context) RemoteAddr() string {
	if c.Conn == nil {
		return ""
	}
	addr := c.Conn.RemoteAddr()
	if addr == nil {
		return ""
	}
	return addr.String()
}
