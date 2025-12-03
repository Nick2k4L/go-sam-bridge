// Package session provides base session implementation.
package session

import (
	"sync"

	i2cp "github.com/go-i2p/go-i2cp"
	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// BaseSession implements the core Session interface functionality.
// All session types (STREAM, DATAGRAM, RAW, PRIMARY) embed BaseSession
// to inherit common lifecycle and state management.
//
// Design: Uses composition pattern per PLAN.md section 6.1
// Concurrency: Protected by RWMutex for thread-safe operations
type BaseSession struct {
	id          string
	config      *SessionConfig
	destination *Destination
	i2cpSession *i2cp.Session
	status      SessionStatus
	mu          sync.RWMutex
}

// NewBaseSession creates a new base session with the given configuration.
// This is typically called by session factory functions, not directly by clients.
func NewBaseSession(config *SessionConfig, dest *Destination, i2cpSess *i2cp.Session) *BaseSession {
	return &BaseSession{
		id:          config.ID,
		config:      config,
		destination: dest,
		i2cpSession: i2cpSess,
		status:      StatusCreating,
	}
}

// ID returns the unique session identifier
func (b *BaseSession) ID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.id
}

// Destination returns the I2P destination for this session
func (b *BaseSession) Destination() *Destination {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.destination
}

// Close terminates the session and releases resources.
// This is the base implementation; specific session types should
// override this to clean up type-specific resources.
func (b *BaseSession) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if already closed
	if b.status == StatusClosed {
		return util.ErrSessionClosed
	}

	// Transition to closing state
	b.status = StatusClosing

	// Close I2CP session if present
	if b.i2cpSession != nil {
		if err := b.i2cpSession.Close(); err != nil {
			b.status = StatusError
			return util.NewSessionError(b.id, "close", err)
		}
	}

	// Mark as closed
	b.status = StatusClosed
	return nil
}

// IsClosed returns true if the session has been closed
func (b *BaseSession) IsClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status == StatusClosed
}

// I2CPSession returns the underlying I2CP session
func (b *BaseSession) I2CPSession() *i2cp.Session {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.i2cpSession
}

// Config returns the session configuration
func (b *BaseSession) Config() *SessionConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// Style returns the session style from configuration
func (b *BaseSession) Style() SessionStyle {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config.Style
}

// Status returns the current session state
func (b *BaseSession) Status() SessionStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// SetStatus updates the session status (protected helper for session implementations)
func (b *BaseSession) SetStatus(status SessionStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

// SetI2CPSession sets the I2CP session (used during initialization)
func (b *BaseSession) SetI2CPSession(sess *i2cp.Session) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.i2cpSession = sess
}
