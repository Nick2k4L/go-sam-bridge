package handler

import (
	"errors"
	"strings"
	"testing"

	commondest "github.com/go-i2p/common/destination"
	"github.com/go-i2p/go-sam-bridge/lib/protocol"
	"github.com/go-i2p/go-sam-bridge/lib/session"
	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// mockSessionRegistry implements session.Registry for testing
type mockSessionRegistry struct {
	sessions    map[string]session.Session
	registerErr error
}

func newMockRegistry() *mockSessionRegistry {
	return &mockSessionRegistry{
		sessions: make(map[string]session.Session),
	}
}

func (r *mockSessionRegistry) Register(s session.Session) error {
	if r.registerErr != nil {
		return r.registerErr
	}
	if _, exists := r.sessions[s.ID()]; exists {
		return util.ErrDuplicateID
	}
	r.sessions[s.ID()] = s
	return nil
}

func (r *mockSessionRegistry) Unregister(id string) error {
	if _, exists := r.sessions[id]; !exists {
		return util.ErrSessionNotFound
	}
	delete(r.sessions, id)
	return nil
}

func (r *mockSessionRegistry) Get(id string) session.Session {
	return r.sessions[id]
}

func (r *mockSessionRegistry) GetByDestination(destHash string) session.Session {
	return nil
}

func (r *mockSessionRegistry) All() []string {
	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	return ids
}

func (r *mockSessionRegistry) Count() int {
	return len(r.sessions)
}

func (r *mockSessionRegistry) Close() error {
	r.sessions = make(map[string]session.Session)
	return nil
}

func TestSessionHandler_Handle(t *testing.T) {
	mockDest := &commondest.Destination{}
	mockPrivKey := []byte("test-private-key")

	successManager := &mockManager{
		dest:        mockDest,
		privateKey:  mockPrivKey,
		pubEncoded:  "test-pub-base64",
		privEncoded: "test-priv-base64",
	}

	tests := []struct {
		name          string
		command       *protocol.Command
		manager       *mockManager
		registry      *mockSessionRegistry
		handshakeDone bool
		sessionBound  bool
		wantResult    string
		wantSession   bool
	}{
		{
			name: "successful STREAM session with TRANSIENT",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultOK,
			wantSession:   true,
		},
		{
			name: "successful with Ed25519 signature type",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":          "STREAM",
					"ID":             "test-session-2",
					"DESTINATION":    "TRANSIENT",
					"SIGNATURE_TYPE": "7",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultOK,
			wantSession:   true,
		},
		{
			name: "missing handshake",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: false,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "session already bound",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			sessionBound:  true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "missing STYLE",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "invalid STYLE",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "INVALID",
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "missing ID",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "ID with whitespace",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "test session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "missing DESTINATION",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE": "STREAM",
					"ID":    "test-session",
				},
			},
			manager:       successManager,
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultI2PError,
		},
		{
			name: "duplicate session ID",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "existing-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager: successManager,
			registry: func() *mockSessionRegistry {
				reg := newMockRegistry()
				reg.registerErr = util.ErrDuplicateID
				return reg
			}(),
			handshakeDone: true,
			wantResult:    protocol.ResultDuplicatedID,
		},
		{
			name: "duplicate destination",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "new-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager: successManager,
			registry: func() *mockSessionRegistry {
				reg := newMockRegistry()
				reg.registerErr = util.ErrDuplicateDest
				return reg
			}(),
			handshakeDone: true,
			wantResult:    protocol.ResultDuplicatedDest,
		},
		{
			name: "key generation failure",
			command: &protocol.Command{
				Verb:   "SESSION",
				Action: "CREATE",
				Options: map[string]string{
					"STYLE":       "STREAM",
					"ID":          "test-session",
					"DESTINATION": "TRANSIENT",
				},
			},
			manager: &mockManager{
				generateErr: errors.New("generation failed"),
			},
			registry:      newMockRegistry(),
			handshakeDone: true,
			wantResult:    protocol.ResultInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSessionHandler(tt.manager)
			ctx := NewContext(&mockConn{}, tt.registry)
			ctx.HandshakeComplete = tt.handshakeDone

			if tt.sessionBound {
				// Simulate already bound session
				ctx.Session = session.NewBaseSession("existing", session.StyleStream, nil, nil, nil)
			}

			resp, err := handler.Handle(ctx, tt.command)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Handle() returned nil response")
			}

			respStr := resp.String()
			if !strings.Contains(respStr, "RESULT="+tt.wantResult) {
				t.Errorf("Handle() = %q, want RESULT=%s", respStr, tt.wantResult)
			}

			if tt.wantSession && ctx.Session == nil {
				t.Error("Handle() did not bind session")
			}

			if tt.wantSession && !strings.Contains(respStr, "DESTINATION=") {
				t.Errorf("Handle() = %q, want DESTINATION=", respStr)
			}
		})
	}
}

func TestSessionHandler_ParseConfig(t *testing.T) {
	handler := NewSessionHandler(&mockManager{})

	tests := []struct {
		name    string
		options map[string]string
		check   func(*session.SessionConfig) bool
	}{
		{
			name:    "defaults",
			options: map[string]string{},
			check: func(c *session.SessionConfig) bool {
				return c.InboundQuantity == 3 && c.OutboundQuantity == 3
			},
		},
		{
			name: "custom tunnel quantities",
			options: map[string]string{
				"inbound.quantity":  "5",
				"outbound.quantity": "5",
			},
			check: func(c *session.SessionConfig) bool {
				return c.InboundQuantity == 5 && c.OutboundQuantity == 5
			},
		},
		{
			name: "custom tunnel lengths",
			options: map[string]string{
				"inbound.length":  "2",
				"outbound.length": "4",
			},
			check: func(c *session.SessionConfig) bool {
				return c.InboundLength == 2 && c.OutboundLength == 4
			},
		},
		{
			name: "port options",
			options: map[string]string{
				"FROM_PORT": "1234",
				"TO_PORT":   "5678",
			},
			check: func(c *session.SessionConfig) bool {
				return c.FromPort == 1234 && c.ToPort == 5678
			},
		},
		{
			name: "RAW options",
			options: map[string]string{
				"PROTOCOL": "20",
				"HEADER":   "true",
			},
			check: func(c *session.SessionConfig) bool {
				return c.Protocol == 20 && c.HeaderEnabled
			},
		},
		{
			name: "invalid values ignored",
			options: map[string]string{
				"inbound.quantity": "invalid",
				"FROM_PORT":        "notaport",
			},
			check: func(c *session.SessionConfig) bool {
				return c.InboundQuantity == 3 && c.FromPort == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &protocol.Command{
				Verb:    "SESSION",
				Action:  "CREATE",
				Options: tt.options,
			}

			config := handler.parseConfig(cmd)
			if !tt.check(config) {
				t.Errorf("parseConfig() returned unexpected config")
			}
		})
	}
}

func TestContainsWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", false},
		{"hello world", true},
		{"hello\tworld", true},
		{"hello\nworld", true},
		{"hello\rworld", true},
		{"", false},
		{" ", true},
		{"  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := containsWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("containsWhitespace(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSessionResponses(t *testing.T) {
	t.Run("sessionOK", func(t *testing.T) {
		resp := sessionOK("test-dest")
		got := resp.String()
		if !strings.Contains(got, "SESSION STATUS") {
			t.Errorf("sessionOK() = %q, want 'SESSION STATUS'", got)
		}
		if !strings.Contains(got, "RESULT=OK") {
			t.Errorf("sessionOK() = %q, want 'RESULT=OK'", got)
		}
		if !strings.Contains(got, "DESTINATION=test-dest") {
			t.Errorf("sessionOK() = %q, want 'DESTINATION=test-dest'", got)
		}
	})

	t.Run("sessionDuplicatedID", func(t *testing.T) {
		resp := sessionDuplicatedID()
		got := resp.String()
		if !strings.Contains(got, "RESULT=DUPLICATED_ID") {
			t.Errorf("sessionDuplicatedID() = %q, want 'RESULT=DUPLICATED_ID'", got)
		}
	})

	t.Run("sessionDuplicatedDest", func(t *testing.T) {
		resp := sessionDuplicatedDest()
		got := resp.String()
		if !strings.Contains(got, "RESULT=DUPLICATED_DEST") {
			t.Errorf("sessionDuplicatedDest() = %q, want 'RESULT=DUPLICATED_DEST'", got)
		}
	})

	t.Run("sessionInvalidKey", func(t *testing.T) {
		resp := sessionInvalidKey("bad key")
		got := resp.String()
		if !strings.Contains(got, "RESULT=INVALID_KEY") {
			t.Errorf("sessionInvalidKey() = %q, want 'RESULT=INVALID_KEY'", got)
		}
	})

	t.Run("sessionError", func(t *testing.T) {
		resp := sessionError("test error")
		got := resp.String()
		if !strings.Contains(got, "RESULT=I2P_ERROR") {
			t.Errorf("sessionError() = %q, want 'RESULT=I2P_ERROR'", got)
		}
		if !strings.Contains(got, "MESSAGE=") {
			t.Errorf("sessionError() = %q, want 'MESSAGE='", got)
		}
	})
}

func TestSessionErr(t *testing.T) {
	err := &sessionErr{msg: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}
