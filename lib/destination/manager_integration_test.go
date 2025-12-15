package destination

import (
	"bytes"
	"testing"

	"github.com/go-i2p/go-sam-bridge/lib/session"
	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// TestGenerateDestination tests Ed25519 destination generation
func TestGenerateDestination(t *testing.T) {
	m := NewManager()

	t.Run("successful generation with Ed25519", func(t *testing.T) {
		dest, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if dest == nil {
			t.Fatal("Generate() returned nil destination")
		}

		// Verify PublicKey is 256 bytes (unused encryption key with padding)
		if len(dest.PublicKey) != 256 {
			t.Errorf("PublicKey length = %d, want 256", len(dest.PublicKey))
		}

		// Verify SigningPrivateKey is 64 bytes (Ed25519)
		if len(dest.SigningPrivateKey) != 64 {
			t.Errorf("SigningPrivateKey length = %d, want 64", len(dest.SigningPrivateKey))
		}

		// Verify PublicKey is not all zeros (has random padding)
		allZeros := true
		for _, b := range dest.PublicKey {
			if b != 0 {
				allZeros = false
				break
			}
		}
		if allZeros {
			t.Error("PublicKey should contain random padding, not all zeros")
		}

		// Verify SigningPrivateKey is not all zeros
		allZeros = true
		for _, b := range dest.SigningPrivateKey {
			if b != 0 {
				allZeros = false
				break
			}
		}
		if allZeros {
			t.Error("SigningPrivateKey should not be all zeros")
		}

		// Verify compressibility of PublicKey (repeating 32-byte pattern)
		// First 32 bytes should repeat 8 times
		pattern := dest.PublicKey[0:32]
		for i := 1; i < 8; i++ {
			segment := dest.PublicKey[i*32 : (i+1)*32]
			if !bytes.Equal(pattern, segment) {
				t.Errorf("PublicKey segment %d does not match pattern (not compressible)", i)
			}
		}
	})

	t.Run("unsupported signature type", func(t *testing.T) {
		_, err := m.Generate(0) // DSA-SHA1
		if err == nil {
			t.Error("Generate() should reject unsupported signature types")
		}
	})

	t.Run("generate multiple destinations - should be unique", func(t *testing.T) {
		dest1, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		dest2, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		// Keys should be different (probabilistically guaranteed)
		if bytes.Equal(dest1.SigningPrivateKey, dest2.SigningPrivateKey) {
			t.Error("Generated destinations should have unique keys")
		}
	})
}

// TestToBase64AndParse tests encoding and parsing roundtrip
func TestToBase64AndParse(t *testing.T) {
	m := NewManager()

	t.Run("roundtrip encode and decode", func(t *testing.T) {
		// Generate a destination
		original, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		// Encode to Base64
		encoded, err := m.ToBase64(original)
		if err != nil {
			t.Fatalf("ToBase64() failed: %v", err)
		}

		// Verify encoded string is not empty
		if encoded == "" {
			t.Error("ToBase64() returned empty string")
		}

		// Decode from Base64
		decoded, err := m.Parse(encoded)
		if err != nil {
			t.Fatalf("Parse() failed: %v", err)
		}

		// Verify roundtrip: keys should match
		if !bytes.Equal(original.PublicKey, decoded.PublicKey) {
			t.Error("Roundtrip failed: PublicKey mismatch")
		}

		if !bytes.Equal(original.SigningPrivateKey, decoded.SigningPrivateKey) {
			t.Error("Roundtrip failed: SigningPrivateKey mismatch")
		}
	})

	t.Run("ToBase64 with nil destination", func(t *testing.T) {
		_, err := m.ToBase64(nil)
		if err != util.ErrInvalidDestination {
			t.Errorf("ToBase64(nil) error = %v, want ErrInvalidDestination", err)
		}
	})

	t.Run("ToBase64 with invalid PublicKey size", func(t *testing.T) {
		invalid := &session.Destination{
			PublicKey:         []byte{0x00}, // Wrong size
			SigningPrivateKey: make([]byte, 64),
		}
		_, err := m.ToBase64(invalid)
		if err == nil {
			t.Error("ToBase64() should reject invalid PublicKey size")
		}
	})

	t.Run("ToBase64 with invalid SigningPrivateKey size", func(t *testing.T) {
		invalid := &session.Destination{
			PublicKey:         make([]byte, 256),
			SigningPrivateKey: []byte{0x00}, // Wrong size
		}
		_, err := m.ToBase64(invalid)
		if err == nil {
			t.Error("ToBase64() should reject invalid SigningPrivateKey size")
		}
	})

	t.Run("Parse empty string", func(t *testing.T) {
		_, err := m.Parse("")
		if err != util.ErrInvalidDestination {
			t.Errorf("Parse(\"\") error = %v, want ErrInvalidDestination", err)
		}
	})

	t.Run("Parse invalid Base64", func(t *testing.T) {
		_, err := m.Parse("!!!invalid!!!")
		if err == nil {
			t.Error("Parse() should reject invalid Base64")
		}
	})

	t.Run("Parse insufficient data", func(t *testing.T) {
		// Valid Base64 but too short
		shortData := util.EncodeDestinationBase64(make([]byte, 100))
		_, err := m.Parse(shortData)
		if err == nil {
			t.Error("Parse() should reject data that's too short")
		}
	})
}

// TestCreateOfflineSignature tests offline signature creation (currently a stub)
func TestCreateOfflineSignature(t *testing.T) {
	m := NewManager()

	t.Run("returns ErrNotImplemented", func(t *testing.T) {
		dest, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		_, err = m.CreateOfflineSignature(dest, 1234567890, 7)
		if err != util.ErrNotImplemented {
			t.Errorf("CreateOfflineSignature() error = %v, want ErrNotImplemented", err)
		}
	})

	t.Run("nil destination check", func(t *testing.T) {
		_, err := m.CreateOfflineSignature(nil, 1234567890, 7)
		if err != util.ErrInvalidDestination {
			t.Errorf("CreateOfflineSignature(nil) error = %v, want ErrInvalidDestination", err)
		}
	})

	t.Run("unsupported signature type check", func(t *testing.T) {
		dest, err := m.Generate(7)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		_, err = m.CreateOfflineSignature(dest, 1234567890, 1) // ECDSA P256
		if err == nil {
			t.Error("CreateOfflineSignature() should reject unsupported signature types")
		}
	})
}

// TestDestinationsEqualWithGeneratedKeys tests the comparison function
func TestDestinationsEqualWithGeneratedKeys(t *testing.T) {
	m := NewManager()

	dest1, err := m.Generate(7)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	dest2, err := m.Generate(7)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Different destinations should not be equal
	if DestinationsEqual(dest1, dest2) {
		t.Error("Different generated destinations should not be equal")
	}

	// Same destination should equal itself
	if !DestinationsEqual(dest1, dest1) {
		t.Error("Same destination should equal itself")
	}
}

// BenchmarkGenerate benchmarks destination generation
func BenchmarkGenerate(b *testing.B) {
	m := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := m.Generate(7)
		if err != nil {
			b.Fatalf("Generate() failed: %v", err)
		}
	}
}

// BenchmarkToBase64 benchmarks encoding
func BenchmarkToBase64(b *testing.B) {
	m := NewManager()
	dest, err := m.Generate(7)
	if err != nil {
		b.Fatalf("Generate() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := m.ToBase64(dest)
		if err != nil {
			b.Fatalf("ToBase64() failed: %v", err)
		}
	}
}

// BenchmarkParse benchmarks parsing
func BenchmarkParse(b *testing.B) {
	m := NewManager()
	dest, err := m.Generate(7)
	if err != nil {
		b.Fatalf("Generate() failed: %v", err)
	}

	encoded, err := m.ToBase64(dest)
	if err != nil {
		b.Fatalf("ToBase64() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := m.Parse(encoded)
		if err != nil {
			b.Fatalf("Parse() failed: %v", err)
		}
	}
}
