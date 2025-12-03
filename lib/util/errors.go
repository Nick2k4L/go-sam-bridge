// Package util provides common utilities for error handling, validation, and encoding.
package util

import (
	"errors"
	"fmt"
)

// Common error types for SAM bridge operations
var (
	// ErrDuplicateID indicates a session ID is already in use
	ErrDuplicateID = errors.New("session ID already exists")

	// ErrDuplicateDestination indicates a destination is already in use by another session
	ErrDuplicateDestination = errors.New("destination already in use")

	// ErrSessionNotFound indicates the requested session does not exist
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionClosed indicates the session is already closed
	ErrSessionClosed = errors.New("session is closed")

	// ErrInvalidSessionStyle indicates an unsupported session style
	ErrInvalidSessionStyle = errors.New("invalid session style")

	// ErrInvalidSignatureType indicates an unsupported signature type
	ErrInvalidSignatureType = errors.New("invalid signature type")

	// ErrInvalidDestination indicates a malformed destination string
	ErrInvalidDestination = errors.New("invalid destination format")

	// ErrInvalidProtocol indicates a disallowed protocol number
	ErrInvalidProtocol = errors.New("protocol number not allowed")

	// ErrInvalidConfiguration indicates invalid session configuration
	ErrInvalidConfiguration = errors.New("invalid session configuration")

	// ErrNotImplemented indicates functionality not yet implemented
	ErrNotImplemented = errors.New("feature not implemented")

	// ErrI2CPConnection indicates I2CP connection failure
	ErrI2CPConnection = errors.New("I2CP connection error")

	// ErrDatagramTooLarge indicates datagram exceeds maximum size
	ErrDatagramTooLarge = errors.New("datagram exceeds maximum size")

	// ErrSubsessionNotAllowed indicates subsession operation on non-PRIMARY session
	ErrSubsessionNotAllowed = errors.New("subsessions only allowed for PRIMARY sessions")
)

// SessionError wraps session-specific errors with context
type SessionError struct {
	SessionID string
	Op        string // Operation being performed
	Err       error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session %s: %s: %v", e.SessionID, e.Op, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

// NewSessionError creates a new SessionError
func NewSessionError(sessionID, op string, err error) *SessionError {
	return &SessionError{
		SessionID: sessionID,
		Op:        op,
		Err:       err,
	}
}

// ProtocolError represents SAM protocol-level errors
type ProtocolError struct {
	Command string
	Message string
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("protocol error in %s: %s", e.Command, e.Message)
}

// NewProtocolError creates a new ProtocolError
func NewProtocolError(command, message string) *ProtocolError {
	return &ProtocolError{
		Command: command,
		Message: message,
	}
}
