package util

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	// Verify all sentinel errors are defined and unique
	sentinels := []error{
		ErrDuplicateID,
		ErrDuplicateDest,
		ErrSessionNotFound,
		ErrInvalidKey,
		ErrTimeout,
		ErrCantReachPeer,
		ErrPeerNotFound,
		ErrLeasesetNotFound,
		ErrKeyNotFound,
		ErrNoVersion,
		ErrAuthRequired,
		ErrAuthFailed,
		ErrSessionClosed,
		ErrNotImplemented,
	}

	for i, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error %d is nil", i)
		}
		if err.Error() == "" {
			t.Errorf("sentinel error %d has empty message", i)
		}
	}

	// Verify errors are unique
	seen := make(map[string]bool)
	for _, err := range sentinels {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("duplicate error message: %q", msg)
		}
		seen[msg] = true
	}
}

func TestSessionError(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		operation string
		err       error
		wantMsg   string
	}{
		{
			name:      "with all fields",
			sessionID: "test123",
			operation: "connect",
			err:       ErrTimeout,
			wantMsg:   "session test123: connect: timeout",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			operation: "accept",
			err:       ErrCantReachPeer,
			wantMsg:   "accept: can't reach peer",
		},
		{
			name:      "with custom error",
			sessionID: "sess1",
			operation: "forward",
			err:       errors.New("custom error"),
			wantMsg:   "session sess1: forward: custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSessionError(tt.sessionID, tt.operation, tt.err)

			// Test Error()
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}

			// Test fields
			if err.SessionID != tt.sessionID {
				t.Errorf("SessionID = %q, want %q", err.SessionID, tt.sessionID)
			}
			if err.Operation != tt.operation {
				t.Errorf("Operation = %q, want %q", err.Operation, tt.operation)
			}

			// Test Unwrap
			if !errors.Is(err, tt.err) {
				t.Error("Unwrap failed: errors.Is returned false")
			}
		})
	}
}

func TestSessionError_Unwrap(t *testing.T) {
	inner := ErrTimeout
	err := NewSessionError("test", "connect", inner)

	// Test errors.Is
	if !errors.Is(err, ErrTimeout) {
		t.Error("errors.Is should find ErrTimeout")
	}
	if errors.Is(err, ErrInvalidKey) {
		t.Error("errors.Is should not find ErrInvalidKey")
	}

	// Test errors.As
	var sessErr *SessionError
	if !errors.As(err, &sessErr) {
		t.Error("errors.As should find SessionError")
	}
	if sessErr.SessionID != "test" {
		t.Errorf("SessionID = %q, want %q", sessErr.SessionID, "test")
	}
}

func TestProtocolError(t *testing.T) {
	tests := []struct {
		name    string
		verb    string
		action  string
		message string
		err     error
		wantMsg string
	}{
		{
			name:    "with verb and action",
			verb:    "SESSION",
			action:  "CREATE",
			message: "invalid style",
			wantMsg: "SESSION CREATE: invalid style",
		},
		{
			name:    "verb only",
			verb:    "PING",
			action:  "",
			message: "unexpected data",
			wantMsg: "PING: unexpected data",
		},
		{
			name:    "with underlying error",
			verb:    "STREAM",
			action:  "CONNECT",
			message: "connection failed",
			err:     ErrCantReachPeer,
			wantMsg: "STREAM CONNECT: connection failed: can't reach peer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err *ProtocolError
			if tt.err != nil {
				err = NewProtocolErrorWithCause(tt.verb, tt.action, tt.message, tt.err)
			} else {
				err = NewProtocolError(tt.verb, tt.action, tt.message)
			}

			// Test Error()
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}

			// Test fields
			if err.Verb != tt.verb {
				t.Errorf("Verb = %q, want %q", err.Verb, tt.verb)
			}
			if err.Action != tt.action {
				t.Errorf("Action = %q, want %q", err.Action, tt.action)
			}
			if err.Message != tt.message {
				t.Errorf("Message = %q, want %q", err.Message, tt.message)
			}

			// Test Unwrap
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Error("Unwrap failed: errors.Is returned false")
			}
		})
	}
}

func TestProtocolError_Unwrap(t *testing.T) {
	inner := ErrInvalidKey
	err := NewProtocolErrorWithCause("SESSION", "CREATE", "bad key", inner)

	// Test errors.Is
	if !errors.Is(err, ErrInvalidKey) {
		t.Error("errors.Is should find ErrInvalidKey")
	}

	// Test errors.As
	var protoErr *ProtocolError
	if !errors.As(err, &protoErr) {
		t.Error("errors.As should find ProtocolError")
	}
	if protoErr.Verb != "SESSION" {
		t.Errorf("Verb = %q, want SESSION", protoErr.Verb)
	}
}

func TestConnectionError(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		operation  string
		err        error
		wantMsg    string
	}{
		{
			name:       "with remote address",
			remoteAddr: "192.168.1.1:7656",
			operation:  "read",
			err:        errors.New("connection reset"),
			wantMsg:    "[192.168.1.1:7656] read: connection reset",
		},
		{
			name:       "empty remote address",
			remoteAddr: "",
			operation:  "write",
			err:        errors.New("broken pipe"),
			wantMsg:    "write: broken pipe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConnectionError(tt.remoteAddr, tt.operation, tt.err)

			// Test Error()
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}

			// Test fields
			if err.RemoteAddr != tt.remoteAddr {
				t.Errorf("RemoteAddr = %q, want %q", err.RemoteAddr, tt.remoteAddr)
			}
			if err.Operation != tt.operation {
				t.Errorf("Operation = %q, want %q", err.Operation, tt.operation)
			}

			// Test Unwrap
			if !errors.Is(err, tt.err) {
				t.Error("Unwrap failed: errors.Is returned false")
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{ErrTimeout, true},
		{ErrCantReachPeer, true},
		{ErrLeasesetNotFound, true},
		{ErrInvalidKey, false},
		{ErrDuplicateID, false},
		{ErrAuthFailed, false},
		{errors.New("unknown error"), false},
		// Wrapped errors
		{NewSessionError("test", "connect", ErrTimeout), true},
		{fmt.Errorf("wrapped: %w", ErrCantReachPeer), true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.err), func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestIsPermanent(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{ErrInvalidKey, true},
		{ErrDuplicateID, true},
		{ErrDuplicateDest, true},
		{ErrAuthFailed, true},
		{ErrNoVersion, true},
		{ErrTimeout, false},
		{ErrCantReachPeer, false},
		{errors.New("unknown error"), false},
		// Wrapped errors
		{NewSessionError("test", "create", ErrDuplicateID), true},
		{fmt.Errorf("wrapped: %w", ErrInvalidKey), true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.err), func(t *testing.T) {
			got := IsPermanent(tt.err)
			if got != tt.expected {
				t.Errorf("IsPermanent(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestToResultCode(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{nil, "OK"},
		{ErrDuplicateID, "DUPLICATED_ID"},
		{ErrDuplicateDest, "DUPLICATED_DEST"},
		{ErrSessionNotFound, "INVALID_ID"},
		{ErrInvalidKey, "INVALID_KEY"},
		{ErrTimeout, "TIMEOUT"},
		{ErrCantReachPeer, "CANT_REACH_PEER"},
		{ErrPeerNotFound, "PEER_NOT_FOUND"},
		{ErrLeasesetNotFound, "LEASESET_NOT_FOUND"},
		{ErrKeyNotFound, "KEY_NOT_FOUND"},
		{ErrNoVersion, "NOVERSION"},
		{errors.New("unknown error"), "I2P_ERROR"},
		// Wrapped errors
		{NewSessionError("test", "op", ErrTimeout), "TIMEOUT"},
		{fmt.Errorf("wrapped: %w", ErrInvalidKey), "INVALID_KEY"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.err), func(t *testing.T) {
			got := ToResultCode(tt.err)
			if got != tt.expected {
				t.Errorf("ToResultCode(%v) = %q, want %q", tt.err, got, tt.expected)
			}
		})
	}
}

func TestErrorWrappingChain(t *testing.T) {
	// Test a multi-level error wrapping chain
	inner := ErrTimeout
	sessionErr := NewSessionError("sess1", "connect", inner)
	protoErr := NewProtocolErrorWithCause("STREAM", "CONNECT", "failed", sessionErr)

	// Should be able to find all errors in the chain
	if !errors.Is(protoErr, ErrTimeout) {
		t.Error("should find ErrTimeout in chain")
	}

	var sessErr *SessionError
	if !errors.As(protoErr, &sessErr) {
		t.Error("should find SessionError in chain")
	}
	if sessErr.SessionID != "sess1" {
		t.Errorf("SessionID = %q, want sess1", sessErr.SessionID)
	}

	// ToResultCode should work through the chain
	if code := ToResultCode(protoErr); code != "TIMEOUT" {
		t.Errorf("ToResultCode = %q, want TIMEOUT", code)
	}
}
