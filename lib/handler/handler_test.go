package handler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	remoteAddr net.Addr
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return m.remoteAddr }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// mockAddr implements net.Addr for testing
type mockAddr struct {
	network string
	addr    string
}

func (m *mockAddr) Network() string { return m.network }
func (m *mockAddr) String() string  { return m.addr }

func TestNewContext(t *testing.T) {
	conn := &mockConn{remoteAddr: &mockAddr{network: "tcp", addr: "127.0.0.1:12345"}}

	ctx := NewContext(conn, nil)

	if ctx.Conn != conn {
		t.Error("Conn not set correctly")
	}
	if ctx.Session != nil {
		t.Error("Session should be nil initially")
	}
	if ctx.Version != "" {
		t.Error("Version should be empty initially")
	}
	if ctx.Authenticated {
		t.Error("Authenticated should be false initially")
	}
	if ctx.HandshakeComplete {
		t.Error("HandshakeComplete should be false initially")
	}
	if ctx.Ctx == nil {
		t.Error("Ctx should have default context")
	}
}

func TestContext_RemoteAddr(t *testing.T) {
	tests := []struct {
		name     string
		conn     net.Conn
		expected string
	}{
		{
			name:     "nil connection",
			conn:     nil,
			expected: "",
		},
		{
			name:     "connection with address",
			conn:     &mockConn{remoteAddr: &mockAddr{network: "tcp", addr: "192.168.1.1:8080"}},
			expected: "192.168.1.1:8080",
		},
		{
			name:     "connection with nil address",
			conn:     &mockConn{remoteAddr: nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{Conn: tt.conn}
			if got := ctx.RemoteAddr(); got != tt.expected {
				t.Errorf("RemoteAddr() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestContext_WithContext(t *testing.T) {
	conn := &mockConn{}
	ctx := NewContext(conn, nil)
	ctx.Version = "3.3"
	ctx.Authenticated = true

	newStdCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newCtx := ctx.WithContext(newStdCtx)

	// Verify the new context has the updated context.Context
	if newCtx.Ctx != newStdCtx {
		t.Error("WithContext did not set the new context")
	}

	// Verify other fields are preserved
	if newCtx.Version != "3.3" {
		t.Error("Version not preserved")
	}
	if !newCtx.Authenticated {
		t.Error("Authenticated not preserved")
	}
	if newCtx.Conn != conn {
		t.Error("Conn not preserved")
	}

	// Verify original context is unchanged
	if ctx.Ctx == newStdCtx {
		t.Error("Original context should not be modified")
	}
}

func TestContext_BindUnbindSession(t *testing.T) {
	ctx := NewContext(nil, nil)

	if ctx.Session != nil {
		t.Error("Session should be nil initially")
	}

	// We can't easily test with a real session without circular imports,
	// but we can test the nil handling
	ctx.BindSession(nil)
	if ctx.Session != nil {
		t.Error("BindSession(nil) should set Session to nil")
	}

	ctx.UnbindSession()
	if ctx.Session != nil {
		t.Error("UnbindSession should set Session to nil")
	}
}

func TestHandlerFunc(t *testing.T) {
	called := false
	fn := HandlerFunc(func(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
		called = true
		return nil, nil
	})

	// Type assertion to verify it implements Handler interface
	var _ Handler = fn

	_, _ = fn.Handle(nil, nil)
	if !called {
		t.Error("HandlerFunc was not called")
	}
}
