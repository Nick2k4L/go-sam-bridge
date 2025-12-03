package destination

import (
	"testing"

	"github.com/go-i2p/go-sam-bridge/lib/session"
)

func TestNewManager(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.cache == nil {
		t.Error("Manager cache is nil")
	}
}

func TestDestinationsEqual(t *testing.T) {
	d1 := &session.Destination{PublicKey: []byte("abc")}
	d2 := &session.Destination{PublicKey: []byte("abc")}
	d3 := &session.Destination{PublicKey: []byte("xyz")}

	if !DestinationsEqual(d1, d2) {
		t.Error("Equal destinations not detected")
	}

	if DestinationsEqual(d1, d3) {
		t.Error("Different destinations marked equal")
	}

	if DestinationsEqual(nil, d1) {
		t.Error("nil should not equal destination")
	}
}
