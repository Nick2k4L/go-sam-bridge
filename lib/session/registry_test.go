package session

import (
	"sync"
	"testing"

	i2cp "github.com/go-i2p/go-i2cp"
)

// mockSession is a minimal Session implementation for testing
type mockSession struct {
	id   string
	dest *Destination
}

func (m *mockSession) ID() string                 { return m.id }
func (m *mockSession) Destination() *Destination  { return m.dest }
func (m *mockSession) Close() error               { return nil }
func (m *mockSession) IsClosed() bool             { return false }
func (m *mockSession) I2CPSession() *i2cp.Session { return nil }
func (m *mockSession) Config() *SessionConfig     { return &SessionConfig{ID: m.id} }
func (m *mockSession) Style() SessionStyle        { return StyleStream }
func (m *mockSession) Status() SessionStatus      { return StatusReady }

// TestRegistry_Register verifies session registration
func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	session := &mockSession{
		id: "test-session-1",
		dest: &Destination{
			PublicKey: []byte("test-key-1"),
		},
	}

	// Register should succeed
	if err := registry.Register(session); err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Verify session is registered
	if registry.Count() != 1 {
		t.Errorf("Count() = %v, want 1", registry.Count())
	}

	// Get the session
	retrieved, exists := registry.Get("test-session-1")
	if !exists {
		t.Error("Get() could not find registered session")
	}
	if retrieved.ID() != "test-session-1" {
		t.Errorf("Retrieved session ID = %v, want test-session-1", retrieved.ID())
	}
}

// TestRegistry_DuplicateID verifies duplicate ID detection
func TestRegistry_DuplicateID(t *testing.T) {
	registry := NewRegistry()

	session1 := &mockSession{id: "duplicate-id", dest: &Destination{PublicKey: []byte("key1")}}
	session2 := &mockSession{id: "duplicate-id", dest: &Destination{PublicKey: []byte("key2")}}

	// First registration should succeed
	if err := registry.Register(session1); err != nil {
		t.Errorf("First Register() error = %v", err)
	}

	// Second registration with same ID should fail
	if err := registry.Register(session2); err == nil {
		t.Error("Register() should fail for duplicate ID")
	}

	// Registry should only have one session
	if registry.Count() != 1 {
		t.Errorf("Count() = %v, want 1 after duplicate attempt", registry.Count())
	}
}

// TestRegistry_DuplicateDestination verifies duplicate destination detection
func TestRegistry_DuplicateDestination(t *testing.T) {
	registry := NewRegistry()

	publicKey := []byte("shared-public-key")
	session1 := &mockSession{
		id:   "session-1",
		dest: &Destination{PublicKey: publicKey},
	}
	session2 := &mockSession{
		id:   "session-2",
		dest: &Destination{PublicKey: publicKey}, // Same public key
	}

	// First registration should succeed
	if err := registry.Register(session1); err != nil {
		t.Errorf("First Register() error = %v", err)
	}

	// Second registration with same destination should fail
	if err := registry.Register(session2); err == nil {
		t.Error("Register() should fail for duplicate destination")
	}

	// Registry should only have one session
	if registry.Count() != 1 {
		t.Errorf("Count() = %v, want 1 after duplicate destination attempt", registry.Count())
	}
}

// TestRegistry_Unregister verifies session removal
func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	session := &mockSession{
		id:   "remove-test",
		dest: &Destination{PublicKey: []byte("remove-key")},
	}

	// Register and verify
	registry.Register(session)
	if registry.Count() != 1 {
		t.Fatal("Setup failed: session not registered")
	}

	// Unregister should succeed
	if err := registry.Unregister("remove-test"); err != nil {
		t.Errorf("Unregister() error = %v", err)
	}

	// Registry should be empty
	if registry.Count() != 0 {
		t.Errorf("Count() = %v, want 0 after Unregister()", registry.Count())
	}

	// Get should not find the session
	_, exists := registry.Get("remove-test")
	if exists {
		t.Error("Get() found session after Unregister()")
	}

	// Second unregister should fail
	if err := registry.Unregister("remove-test"); err == nil {
		t.Error("Second Unregister() should fail")
	}
}

// TestRegistry_List verifies listing all session IDs
func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Register multiple sessions
	sessions := []string{"session-a", "session-b", "session-c"}
	for i, id := range sessions {
		session := &mockSession{
			id:   id,
			dest: &Destination{PublicKey: []byte{byte(i)}},
		}
		registry.Register(session)
	}

	// List should return all IDs
	ids := registry.List()
	if len(ids) != 3 {
		t.Errorf("List() returned %v items, want 3", len(ids))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expected := range sessions {
		if !idMap[expected] {
			t.Errorf("List() missing session %v", expected)
		}
	}
}

// TestRegistry_CheckDuplicateDestination verifies destination checking
func TestRegistry_CheckDuplicateDestination(t *testing.T) {
	registry := NewRegistry()

	dest := &Destination{PublicKey: []byte("check-key")}
	session := &mockSession{
		id:   "check-session",
		dest: dest,
	}

	// Initially should not be duplicate
	if registry.CheckDuplicateDestination(dest) {
		t.Error("CheckDuplicateDestination() should return false before registration")
	}

	// Register session
	registry.Register(session)

	// Now should be duplicate
	if !registry.CheckDuplicateDestination(dest) {
		t.Error("CheckDuplicateDestination() should return true after registration")
	}

	// Different destination should not be duplicate
	differentDest := &Destination{PublicKey: []byte("different-key")}
	if registry.CheckDuplicateDestination(differentDest) {
		t.Error("CheckDuplicateDestination() should return false for different destination")
	}

	// Nil destination should not be duplicate
	if registry.CheckDuplicateDestination(nil) {
		t.Error("CheckDuplicateDestination() should return false for nil destination")
	}
}

// TestRegistry_ConcurrentAccess verifies thread-safety
func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent registrations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			session := &mockSession{
				id:   string(rune('A' + index)),
				dest: &Destination{PublicKey: []byte{byte(index)}},
			}
			if err := registry.Register(session); err != nil {
				errors <- err
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.List()
			registry.Count()
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	// Verify final count
	if count := registry.Count(); count != 50 {
		t.Errorf("Count() = %v, want 50 after concurrent operations", count)
	}
}

// TestRegistry_CloseAll verifies closing all sessions
func TestRegistry_CloseAll(t *testing.T) {
	registry := NewRegistry()

	// Register multiple sessions
	for i := 0; i < 5; i++ {
		session := &mockSession{
			id:   string(rune('A' + i)),
			dest: &Destination{PublicKey: []byte{byte(i)}},
		}
		registry.Register(session)
	}

	if registry.Count() != 5 {
		t.Fatalf("Setup failed: count = %v, want 5", registry.Count())
	}

	// Close all sessions
	if err := registry.CloseAll(); err != nil {
		t.Errorf("CloseAll() error = %v", err)
	}

	// Registry should be empty
	if registry.Count() != 0 {
		t.Errorf("Count() = %v, want 0 after CloseAll()", registry.Count())
	}
}

// TestRegistry_NilSession verifies nil session handling
func TestRegistry_NilSession(t *testing.T) {
	registry := NewRegistry()

	if err := registry.Register(nil); err == nil {
		t.Error("Register(nil) should return error")
	}
}
