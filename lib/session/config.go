// Package session provides SAM session configuration structures.
package session

// SessionConfig holds creation parameters for all session types.
// This structure is used during SESSION CREATE and SESSION ADD commands.
type SessionConfig struct {
	// ID is the unique session identifier
	ID string

	// Style determines the session type (STREAM, DATAGRAM, RAW, etc.)
	Style SessionStyle

	// Destination is the I2P destination; nil for TRANSIENT destinations
	Destination *Destination

	// Signature and encryption configuration
	SignatureType   int   // Default 7 (Ed25519)
	EncryptionTypes []int // Default [4, 0] (ECIES + ElGamal)

	// Port and protocol configuration (SAM 3.2+)
	FromPort       int // Source port for connections
	ToPort         int // Destination port for connections
	Protocol       int // Protocol number (RAW only)
	ListenPort     int // Listening port for PRIMARY subsessions
	ListenProtocol int // Listening protocol for PRIMARY RAW subsessions

	// Datagram forwarding configuration
	ForwardHost string // UDP forwarding host
	ForwardPort int    // UDP forwarding port

	// RAW specific options
	HeaderMode bool // Prepend header to forwarded datagrams (SAM 3.2)

	// I2CP options for tunnel configuration
	I2CPOptions map[string]string

	// Streaming library options
	StreamingOptions map[string]string

	// Offline signature support for enhanced privacy
	OfflineSignature *OfflineSignature
}

// SubsessionConfig extends SessionConfig for PRIMARY session subsessions.
// Subsessions inherit the parent's I2P destination but have independent
// routing and protocol settings.
type SubsessionConfig struct {
	SessionConfig        // Embed base configuration
	ParentID      string // Reference to the PRIMARY session
}

// ConnectOptions configures STREAM CONNECT operations
type ConnectOptions struct {
	Silent   bool // Don't send data immediately (SAM 3.1+)
	FromPort int  // Source port (SAM 3.2+)
	ToPort   int  // Destination port (SAM 3.2+)
}

// ForwardOptions configures STREAM FORWARD operations
type ForwardOptions struct {
	Silent bool // Don't send initial data (SAM 3.1+)
	SSL    bool // Use SSL/TLS for forwarded connection (SAM 3.2+)
}

// DatagramOptions configures datagram sending operations
type DatagramOptions struct {
	FromPort     int  // Source port (SAM 3.2+)
	ToPort       int  // Destination port (SAM 3.2+)
	SendTags     int  // Session tag count override (SAM 3.3)
	TagThreshold int  // Low tag threshold (SAM 3.3)
	Expires      int  // Message expiration in seconds (SAM 3.3)
	SendLeaseSet bool // Bundle lease set with message (SAM 3.3)
}

// RawOptions configures raw datagram sending operations
type RawOptions struct {
	FromPort     int  // Source port (SAM 3.2+)
	ToPort       int  // Destination port (SAM 3.2+)
	Protocol     int  // Protocol number (SAM 3.2+)
	SendTags     int  // Session tag count override (SAM 3.3)
	TagThreshold int  // Low tag threshold (SAM 3.3)
	Expires      int  // Message expiration in seconds (SAM 3.3)
	SendLeaseSet bool // Bundle lease set with message (SAM 3.3)
}

// DatagramMetadata contains metadata received with datagrams
type DatagramMetadata struct {
	FromPort int // Source port
	ToPort   int // Destination port
}

// RawMetadata contains metadata received with raw datagrams
type RawMetadata struct {
	FromPort int // Source port
	ToPort   int // Destination port
	Protocol int // Protocol number
}

// Destination wraps I2P destination key material.
// Contains public and private keys for I2P communication.
type Destination struct {
	// PublicKey is the I2P destination public key
	PublicKey []byte

	// PrivateKey is the encryption private key (may be zeroed)
	PrivateKey []byte

	// SigningPrivateKey is the signing private key
	SigningPrivateKey []byte

	// OfflineSig is optional offline signature for enhanced privacy
	OfflineSig *OfflineSignature
}

// OfflineSignature enables offline signing for STREAM and RAW sessions.
// This allows using a transient key while keeping the long-term identity key offline.
type OfflineSignature struct {
	Expires             uint32 // Expiration timestamp (Unix time)
	TransientSigType    uint16 // Signature type of transient key
	TransientPublicKey  []byte // Transient public signing key
	Signature           []byte // Signature by long-term key
	TransientPrivateKey []byte // Transient private signing key (not transmitted)
}
