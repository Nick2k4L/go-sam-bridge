// Package destination provides I2P destination management and key generation.
package destination

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/go-i2p/crypto/ed25519"
	"github.com/go-i2p/go-sam-bridge/lib/session"
	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// Manager handles destination generation, parsing, and storage.
// Per PLAN.md section 3.2: Supports multiple signature types and offline signatures.
//
// The Manager provides destination caching to avoid regenerating
// the same destinations and supports Base64 encoding/decoding.
type Manager struct {
	cache map[string]*session.Destination
}

// NewManager creates a new destination manager
func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]*session.Destination),
	}
}

// Generate creates a new I2P destination with the specified signature type.
// Supported signature types:
//   - 7: Ed25519 (default, recommended per SAM 3.3 and I2P specs)
//
// The generated destination includes:
//   - PublicKey: Ed25519 public signing key (32 bytes)
//   - PrivateKey: Unused encryption key filled with secure random data (256 bytes, per I2P Proposal 161)
//   - SigningPrivateKey: Ed25519 private signing key (64 bytes)
//
// Per I2P Proposal 161, the encryption PublicKey field in destinations has been
// unused since I2P 0.6 (2005). We fill it with compressible random data for
// backward compatibility and to prevent Base64 addresses from appearing corrupt.
//
// Reference: SAMv3.md DEST GENERATE command, I2P Common Structures specification
func (m *Manager) Generate(sigType int) (*session.Destination, error) {
	// Validate signature type (currently only Ed25519 is supported)
	if sigType != 7 {
		return nil, fmt.Errorf("%w: type %d not supported (only Ed25519/7 currently)", util.ErrInvalidSignatureType, sigType)
	}

	// Generate Ed25519 key pair for signing
	publicKey, privateKey, err := ed25519.GenerateEd25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	_ = publicKey // Public key is embedded in the 64-byte private key (last 32 bytes)

	// Create destination with generated keys
	// Note: I2P Destinations have 3 key fields:
	// - PublicKey (256 bytes): Unused encryption key since I2P 0.6, filled with random padding
	// - PrivateKey: Not stored in destination (encryption private key, unused)
	// - SigningPrivateKey (64 bytes): Ed25519 private key for signing
	dest := &session.Destination{
		PublicKey:         make([]byte, 256),  // Will be filled with random padding below
		SigningPrivateKey: privateKey.Bytes(), // Ed25519 64-byte private key
		PrivateKey:        nil,                // Unused encryption private key
	}

	// Fill PublicKey with secure random data per I2P Proposal 161
	// This makes the KeysAndCert structure highly compressible in I2P protocols
	// (SSU2 handshake, Streaming SYN, Datagram messages, etc.)
	randomData := make([]byte, 32)
	if _, err := rand.Read(randomData); err != nil {
		privateKey.Zero() // Clean up on error
		return nil, fmt.Errorf("failed to generate random padding: %w", err)
	}

	// Repeat 32-byte random pattern to fill 256-byte PublicKey (8 repetitions)
	// This ensures high compressibility while maintaining randomness for entropy
	for i := 0; i < 8; i++ {
		copy(dest.PublicKey[i*32:(i+1)*32], randomData)
	}

	return dest, nil
}

// Parse converts a Base64-encoded private key string to a Destination.
// The private key string format follows the PrivateKeyFile specification:
//   - I2P Base64-encoded binary data (using I2P alphabet: - and ~ instead of + and /)
//   - Contains: PublicKey (256 bytes) + SigningPrivateKey (variable) + Certificate
//   - Minimum size: 387 bytes for Ed25519 (256 + 128 unused + 3 cert)
//
// This parses the SAM DEST GENERATE response format and destination strings.
// Per I2P spec, the 256-byte PublicKey field has been unused since 0.6 (2005).
//
// Reference: SAMv3.md, I2P Common Structures
func (m *Manager) Parse(privKey string) (*session.Destination, error) {
	if privKey == "" {
		return nil, util.ErrInvalidDestination
	}

	// Decode using I2P Base64 alphabet
	data, err := util.DecodeDestinationBase64(privKey)
	if err != nil {
		return nil, err // Error already wrapped
	}

	// Validate minimum length for Ed25519 destination
	// Format: PublicKey (256) + Padding (128) + SigningPublicKey (32) + SigningPrivateKey (64) = 480 bytes minimum
	// However, SAM format may vary - we need at least the signing private key
	if len(data) < 387 {
		return nil, fmt.Errorf("%w: insufficient data length (got %d, need at least 387)", util.ErrInvalidDestination, len(data))
	}

	// Parse basic structure:
	// Bytes 0-255: PublicKey (256 bytes, unused encryption key)
	// Bytes 256-383: Unused padding (128 bytes)
	// Bytes 384+: SigningPrivateKey + Certificate (variable length)
	//
	// For Ed25519: SigningPrivateKey is 64 bytes
	// TODO: Full parsing with certificate support requires go-i2p/common integration
	//       For Phase 1, we extract the keys and create a minimal destination

	dest := &session.Destination{
		PublicKey:         make([]byte, 256),
		SigningPrivateKey: nil, // Will be set below after validation
		PrivateKey:        nil, // Unused encryption private key
	}

	// Copy PublicKey (first 256 bytes)
	copy(dest.PublicKey, data[0:256])

	// For Ed25519, extract the 64-byte signing private key
	// The private key starts after: PublicKey (256) + Padding (128) = byte 384
	if len(data) >= 384+64 {
		dest.SigningPrivateKey = make([]byte, 64)
		copy(dest.SigningPrivateKey, data[384:384+64])
	} else {
		return nil, fmt.Errorf("%w: insufficient data for Ed25519 private key", util.ErrInvalidDestination)
	}

	return dest, nil
}

// ToBase64 converts a Destination to an I2P Base64-encoded private key string.
// This format matches the SAM DEST GENERATE response and can be stored in PrivateKeyFile.
//
// The encoded format includes:
//   - PublicKey (256 bytes): Unused encryption key with random padding
//   - Unused padding (128 bytes): For backward compatibility
//   - SigningPrivateKey (64 bytes for Ed25519)
//   - Certificate data (minimal)
//
// The output uses I2P Base64 alphabet (- and ~ instead of + and /).
//
// Reference: SAMv3.md DEST GENERATE response format
func (m *Manager) ToBase64(dest *session.Destination) (string, error) {
	if dest == nil {
		return "", util.ErrInvalidDestination
	}

	if len(dest.PublicKey) != 256 {
		return "", fmt.Errorf("%w: PublicKey must be 256 bytes", util.ErrInvalidDestination)
	}

	if len(dest.SigningPrivateKey) != 64 {
		return "", fmt.Errorf("%w: SigningPrivateKey must be 64 bytes for Ed25519", util.ErrInvalidDestination)
	}

	// Build binary format matching I2P destination structure:
	// - PublicKey (256 bytes)
	// - Unused padding (128 bytes, typically zeros)
	// - SigningPrivateKey (64 bytes)
	// - Minimal certificate (will be added in future with go-i2p/common integration)

	data := make([]byte, 256+128+64)

	// Copy PublicKey (256 bytes)
	copy(data[0:256], dest.PublicKey)

	// Padding (128 bytes) - left as zeros for now
	// Future: This should match the destination's certificate structure

	// Copy SigningPrivateKey (64 bytes)
	copy(data[384:384+64], dest.SigningPrivateKey)

	// Encode to I2P Base64
	return util.EncodeDestinationBase64(data), nil
}

// CreateOfflineSignature generates an offline signature for the destination.
// Offline signatures enable using a transient key for day-to-day operations
// while keeping the long-term identity key offline for enhanced security.
//
// Parameters:
//   - dest: The destination to create an offline signature for
//   - expires: Unix timestamp when the signature expires
//   - transientSigType: Signature type for the transient key (typically 7 for Ed25519)
//
// Returns an OfflineSignature that can be used with STREAM and RAW sessions.
//
// TODO: Implementation requires:
//  1. Generate transient Ed25519 key pair
//  2. Create signature data: expires || transientSigType || transientPublicKey
//  3. Sign with dest.SigningPrivateKey using Ed25519
//  4. Return OfflineSignature structure
//
// Reference: I2P Offline Signatures specification (proposal 123)
func (m *Manager) CreateOfflineSignature(dest *session.Destination, expires uint32, transientSigType uint16) (*session.OfflineSignature, error) {
	if dest == nil {
		return nil, util.ErrInvalidDestination
	}

	if transientSigType != 7 {
		return nil, fmt.Errorf("%w: only Ed25519 (type 7) supported", util.ErrInvalidSignatureType)
	}

	// TODO: Implement offline signature generation
	// Requires: go-i2p/crypto ed25519 signing integration
	// For Phase 1, this is acceptable as offline signatures are optional
	return nil, util.ErrNotImplemented
}

// DestinationsEqual compares two destinations for equality.
// Two destinations are equal if their public keys match.
func DestinationsEqual(a, b *session.Destination) bool {
	if a == nil || b == nil {
		return false
	}

	return bytes.Equal(a.PublicKey, b.PublicKey)
}
