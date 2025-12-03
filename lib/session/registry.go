// Package session provides session registry for tracking active sessions.
package session

import (
	"sync"

	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// Registry tracks all active SAM sessions globally.
// Per PLAN.md section 3.5: Enforces unique session IDs and destinations.
//
// Concurrency: Thread-safe using RWMutex for concurrent access.
// The registry is shared across all SAM client connections.
type Registry struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

// NewRegistry creates a new session registry
func NewRegistry() *Registry {
	return &Registry{
		sessions: make(map[string]Session),
	}
}

// Register adds a new session to the registry.
// Returns an error if:
// - A session with the same ID already exists
// - A session with the same destination already exists (per SAM spec)
func (r *Registry) Register(session Session) error {
	if session == nil {
		return util.NewProtocolError("Register", "session cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate ID
	if _, exists := r.sessions[session.ID()]; exists {
		return util.ErrDuplicateID
	}

	// Check for duplicate destination (SAM spec requirement)
	dest := session.Destination()
	if dest != nil {
		for _, existing := range r.sessions {
			existingDest := existing.Destination()
			if existingDest != nil && destinationsEqual(dest, existingDest) {
				return util.ErrDuplicateDestination
			}
		}
	}

	r.sessions[session.ID()] = session
	return nil
}

// Unregister removes a session from the registry by ID.
// Returns an error if the session does not exist.
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id]; !exists {
		return util.ErrSessionNotFound
	}

	delete(r.sessions, id)
	return nil
}

// Get retrieves a session by ID.
// Returns (session, true) if found, or (nil, false) if not found.
func (r *Registry) Get(id string) (Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	return session, exists
}

// List returns all active session IDs.
// The returned slice is a snapshot and safe to iterate.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of active sessions
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}

// CheckDuplicateDestination returns true if a session with the given
// destination already exists in the registry.
func (r *Registry) CheckDuplicateDestination(dest *Destination) bool {
	if dest == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		existingDest := session.Destination()
		if existingDest != nil && destinationsEqual(dest, existingDest) {
			return true
		}
	}
	return false
}

// CloseAll closes all registered sessions.
// This is typically called during bridge shutdown.
// Errors are collected and the first error is returned.
func (r *Registry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for id, session := range r.sessions {
		if err := session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(r.sessions, id)
	}

	return firstErr
}

// destinationsEqual compares two destinations for equality.
// Two destinations are equal if their public keys match.
// This is used to enforce SAM spec requirement of unique destinations.
func destinationsEqual(a, b *Destination) bool {
	if a == nil || b == nil {
		return false
	}

	// Compare public keys (the unique identifier for a destination)
	if len(a.PublicKey) != len(b.PublicKey) {
		return false
	}

	for i := range a.PublicKey {
		if a.PublicKey[i] != b.PublicKey[i] {
			return false
		}
	}

	return true
}
