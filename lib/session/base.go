// Package session implements SAM v3.0-3.3 session management.
package session

import (
	"context"
	"net"
	"sync"
)

// BaseSession provides common functionality for all session types.
// All session implementations must embed *BaseSession per project guidelines.
// BaseSession is thread-safe with RWMutex protection for all field access.
type BaseSession struct {
	mu sync.RWMutex

	id          string
	style       Style
	destination *Destination
	status      Status
	controlConn net.Conn
	config      *SessionConfig

	// i2cpSession holds the I2CP session handle for tunnel management.
	// ISSUE-003: Used to wait for tunnel readiness and manage I2CP lifecycle.
	i2cpSession I2CPSessionHandle
}

// NewBaseSession creates a new BaseSession with the given parameters.
// The session starts in StatusCreating state.
func NewBaseSession(id string, style Style, dest *Destination, conn net.Conn, cfg *SessionConfig) *BaseSession {
	if cfg == nil {
		cfg = DefaultSessionConfig()
	}
	return &BaseSession{
		id:          id,
		style:       style,
		destination: dest,
		status:      StatusCreating,
		controlConn: conn,
		config:      cfg,
	}
}

// ID returns the unique session identifier (nickname).
// Session IDs must be globally unique per SAMv3.md.
func (b *BaseSession) ID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.id
}

// Style returns the session style (STREAM, DATAGRAM, RAW, PRIMARY, etc.).
func (b *BaseSession) Style() Style {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.style
}

// Destination returns the I2P destination associated with this session.
func (b *BaseSession) Destination() *Destination {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.destination
}

// Status returns the current session status.
func (b *BaseSession) Status() Status {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// ControlConn returns the control socket associated with this session.
// Session dies when this socket closes per SAMv3.md.
func (b *BaseSession) ControlConn() net.Conn {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.controlConn
}

// Config returns the session configuration.
func (b *BaseSession) Config() *SessionConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// SetStatus updates the session status.
// This is used internally during session lifecycle transitions.
func (b *BaseSession) SetStatus(s Status) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = s
}

// SetDestination updates the session destination.
// This is used after key generation during session creation.
func (b *BaseSession) SetDestination(dest *Destination) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.destination = dest
}

// Activate transitions the session from Creating to Active status.
// Returns false if the session is not in Creating status.
func (b *BaseSession) Activate() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.status != StatusCreating {
		return false
	}
	b.status = StatusActive
	return true
}

// Close terminates the session and releases all resources.
// Close is safe to call multiple times; subsequent calls are no-ops.
// Implements the Session interface Close method.
func (b *BaseSession) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Already closed or closing
	if b.status == StatusClosed || b.status == StatusClosing {
		return nil
	}

	b.status = StatusClosing

	var errs []error

	// Close I2CP session first
	if b.i2cpSession != nil {
		if err := b.i2cpSession.Close(); err != nil {
			errs = append(errs, err)
		}
		b.i2cpSession = nil
	}

	// Close control connection
	if b.controlConn != nil {
		if err := b.controlConn.Close(); err != nil {
			errs = append(errs, err)
		}
		b.controlConn = nil
	}

	b.status = StatusClosed

	if len(errs) > 0 {
		return errs[0] // Return first error
	}
	return nil
}

// IsClosed returns true if the session has been closed.
func (b *BaseSession) IsClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status == StatusClosed
}

// IsActive returns true if the session is currently active.
func (b *BaseSession) IsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status == StatusActive
}

// SetI2CPSession sets the I2CP session handle.
// ISSUE-003: Allows handler to associate I2CP session with SAM session.
func (b *BaseSession) SetI2CPSession(i2cp I2CPSessionHandle) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.i2cpSession = i2cp
}

// I2CPSession returns the I2CP session handle, if set.
func (b *BaseSession) I2CPSession() I2CPSessionHandle {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.i2cpSession
}

// WaitForTunnels blocks until tunnels are built or context is cancelled.
// Per SAMv3.md: "the router builds tunnels before responding with SESSION STATUS.
// This could take several seconds."
// ISSUE-003: Use this to block SESSION STATUS response until tunnels are ready.
// Returns nil immediately if no I2CP session is set.
func (b *BaseSession) WaitForTunnels(ctx context.Context) error {
	b.mu.RLock()
	i2cp := b.i2cpSession
	b.mu.RUnlock()

	if i2cp == nil {
		return nil // No I2CP session, don't block
	}
	return i2cp.WaitForTunnels(ctx)
}
