// Package session implements SAM v3.0-3.3 session management.
package session

import (
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

	var err error
	if b.controlConn != nil {
		err = b.controlConn.Close()
		b.controlConn = nil
	}

	b.status = StatusClosed
	return err
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
