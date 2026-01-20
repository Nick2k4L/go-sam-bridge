package session

import (
	"net"
	"sync"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing.
type mockConn struct {
	closed bool
	mu     sync.Mutex
}

func (m *mockConn) Read(_ []byte) (n int, err error)  { return 0, nil }
func (m *mockConn) Write(_ []byte) (n int, err error) { return 0, nil }
func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(_ time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(_ time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(_ time.Time) error { return nil }

func (m *mockConn) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestNewBaseSession(t *testing.T) {
	dest := &Destination{PublicKey: []byte("test")}
	conn := &mockConn{}
	cfg := DefaultSessionConfig()

	session := NewBaseSession("test-id", StyleStream, dest, conn, cfg)

	if session.ID() != "test-id" {
		t.Errorf("ID() = %q, want %q", session.ID(), "test-id")
	}
	if session.Style() != StyleStream {
		t.Errorf("Style() = %q, want %q", session.Style(), StyleStream)
	}
	if session.Destination() != dest {
		t.Error("Destination() should return the provided destination")
	}
	if session.Status() != StatusCreating {
		t.Errorf("Status() = %v, want StatusCreating", session.Status())
	}
	if session.ControlConn() != conn {
		t.Error("ControlConn() should return the provided connection")
	}
	if session.Config() != cfg {
		t.Error("Config() should return the provided config")
	}
}

func TestNewBaseSession_NilConfig(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	cfg := session.Config()
	if cfg == nil {
		t.Fatal("Config() should return default config when nil provided")
	}
	if cfg.SignatureType != DefaultSignatureType {
		t.Errorf("default config SignatureType = %d, want %d", cfg.SignatureType, DefaultSignatureType)
	}
}

func TestBaseSession_SetStatus(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	if session.Status() != StatusCreating {
		t.Errorf("initial Status() = %v, want StatusCreating", session.Status())
	}

	session.SetStatus(StatusActive)
	if session.Status() != StatusActive {
		t.Errorf("Status() = %v, want StatusActive", session.Status())
	}

	session.SetStatus(StatusClosing)
	if session.Status() != StatusClosing {
		t.Errorf("Status() = %v, want StatusClosing", session.Status())
	}

	session.SetStatus(StatusClosed)
	if session.Status() != StatusClosed {
		t.Errorf("Status() = %v, want StatusClosed", session.Status())
	}
}

func TestBaseSession_SetDestination(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	if session.Destination() != nil {
		t.Error("initial Destination() should be nil")
	}

	dest := &Destination{PublicKey: []byte("newdest")}
	session.SetDestination(dest)

	if session.Destination() != dest {
		t.Error("Destination() should return newly set destination")
	}
}

func TestBaseSession_Activate(t *testing.T) {
	t.Run("activate from creating", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

		if !session.Activate() {
			t.Error("Activate() should return true when in Creating status")
		}
		if session.Status() != StatusActive {
			t.Errorf("Status() = %v, want StatusActive", session.Status())
		}
	})

	t.Run("activate from active fails", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)
		session.SetStatus(StatusActive)

		if session.Activate() {
			t.Error("Activate() should return false when already Active")
		}
	})

	t.Run("activate from closed fails", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)
		session.SetStatus(StatusClosed)

		if session.Activate() {
			t.Error("Activate() should return false when Closed")
		}
	})
}

func TestBaseSession_Close(t *testing.T) {
	t.Run("close with connection", func(t *testing.T) {
		conn := &mockConn{}
		session := NewBaseSession("test-id", StyleStream, nil, conn, nil)
		session.SetStatus(StatusActive)

		err := session.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		if session.Status() != StatusClosed {
			t.Errorf("Status() = %v, want StatusClosed", session.Status())
		}
		if !conn.isClosed() {
			t.Error("Connection should be closed")
		}
	})

	t.Run("close without connection", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)
		session.SetStatus(StatusActive)

		err := session.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		if session.Status() != StatusClosed {
			t.Errorf("Status() = %v, want StatusClosed", session.Status())
		}
	})

	t.Run("close idempotent", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)
		session.SetStatus(StatusActive)

		_ = session.Close()
		err := session.Close()
		if err != nil {
			t.Errorf("second Close() returned error: %v", err)
		}
		if session.Status() != StatusClosed {
			t.Errorf("Status() = %v, want StatusClosed", session.Status())
		}
	})

	t.Run("close already closing", func(t *testing.T) {
		session := NewBaseSession("test-id", StyleStream, nil, nil, nil)
		session.SetStatus(StatusClosing)

		err := session.Close()
		if err != nil {
			t.Errorf("Close() on Closing status returned error: %v", err)
		}
	})
}

func TestBaseSession_IsClosed(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	if session.IsClosed() {
		t.Error("IsClosed() should be false initially")
	}

	session.SetStatus(StatusActive)
	if session.IsClosed() {
		t.Error("IsClosed() should be false when Active")
	}

	session.SetStatus(StatusClosed)
	if !session.IsClosed() {
		t.Error("IsClosed() should be true when Closed")
	}
}

func TestBaseSession_IsActive(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	if session.IsActive() {
		t.Error("IsActive() should be false initially (Creating)")
	}

	session.SetStatus(StatusActive)
	if !session.IsActive() {
		t.Error("IsActive() should be true when Active")
	}

	session.SetStatus(StatusClosing)
	if session.IsActive() {
		t.Error("IsActive() should be false when Closing")
	}

	session.SetStatus(StatusClosed)
	if session.IsActive() {
		t.Error("IsActive() should be false when Closed")
	}
}

func TestBaseSession_ConcurrentAccess(t *testing.T) {
	session := NewBaseSession("test-id", StyleStream, nil, nil, nil)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = session.ID()
				_ = session.Style()
				_ = session.Destination()
				_ = session.Status()
				_ = session.ControlConn()
				_ = session.Config()
				_ = session.IsClosed()
				_ = session.IsActive()
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				session.SetStatus(StatusActive)
				session.SetDestination(&Destination{PublicKey: []byte("test")})
			}
		}()
	}

	wg.Wait()
}
