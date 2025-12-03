// Package protocol defines SAM protocol constants, command structures, and response builders.
package protocol

// SAM protocol version constants
const (
	// MinSAMVersion is the minimum supported SAM protocol version
	MinSAMVersion = "3.0"
	// MaxSAMVersion is the maximum supported SAM protocol version
	MaxSAMVersion = "3.3"
)

// SAM command types
const (
	CmdHello         = "HELLO"
	CmdSessionCreate = "SESSION CREATE"
	CmdSessionAdd    = "SESSION ADD"
	CmdSessionRemove = "SESSION REMOVE"
	CmdStreamConnect = "STREAM CONNECT"
	CmdStreamAccept  = "STREAM ACCEPT"
	CmdStreamForward = "STREAM FORWARD"
	CmdDatagramSend  = "DATAGRAM SEND"
	CmdRawSend       = "RAW SEND"
	CmdDestGenerate  = "DEST GENERATE"
	CmdNamingLookup  = "NAMING LOOKUP"
	CmdPing          = "PING"
	CmdPong          = "PONG"
	CmdQuit          = "QUIT"
	CmdStop          = "STOP"
	CmdExit          = "EXIT"
)

// SAM result codes as defined in the specification
const (
	ResultOK                = "OK"
	ResultCantReachPeer     = "CANT_REACH_PEER"
	ResultDuplicatedDest    = "DUPLICATED_DEST"
	ResultDuplicatedID      = "DUPLICATED_ID"
	ResultI2PError          = "I2P_ERROR"
	ResultInvalidKey        = "INVALID_KEY"
	ResultInvalidID         = "INVALID_ID"
	ResultKeyNotFound       = "KEY_NOT_FOUND"
	ResultPeerNotFound      = "PEER_NOT_FOUND"
	ResultTimeout           = "TIMEOUT"
	ResultAlreadyAccepting  = "ALREADY_ACCEPTING"
	ResultNoVersion         = "NOVERSION"
	ResultBadSyntax         = "BADSYNTAX"
	ResultBadOptions        = "BADOPTIONS"
	ResultBadSessionStyle   = "BADSESSIONSTYLE"
	ResultNotEnoughRam      = "NOTENOUGHRAM"
	ResultSessionAlreadySet = "SESSIONALREADYSET"
)

// Session style constants
const (
	StyleStream    = "STREAM"
	StyleDatagram  = "DATAGRAM"
	StyleRaw       = "RAW"
	StyleDatagram2 = "DATAGRAM2"
	StyleDatagram3 = "DATAGRAM3"
	StylePrimary   = "PRIMARY"
)

// Default I2CP and crypto options as recommended in PLAN.md
const (
	// DefaultSignatureType is Ed25519 (type 7)
	DefaultSignatureType = 7
	// DefaultTunnelQuantity balances performance and resource usage
	DefaultTunnelQuantity = 3
)

// DefaultEncryptionTypes provides ECIES with ElGamal fallback
var DefaultEncryptionTypes = []int{4, 0}

// Protocol numbers for routing (from SAM 3.2+)
const (
	// ProtocolTCP is reserved for streaming (protocol 6)
	ProtocolTCP = 6
	// ProtocolUDP is commonly used for datagrams (protocol 17)
	ProtocolUDP = 17
	// Disallowed protocols: 6 (TCP/streaming), 17 (UDP), 19, 20
)

// DisallowedProtocols lists protocol numbers that cannot be used with RAW sessions
var DisallowedProtocols = []int{6, 17, 19, 20}

// MTU size limits for different datagram types
const (
	// MaxRawDatagramSize is the maximum size for raw datagrams (SAM spec)
	MaxRawDatagramSize = 32768
	// MaxRepliableDatagramSize is the maximum size for repliable datagrams
	MaxRepliableDatagramSize = 31744
	// RecommendedDatagramSize is the recommended safe size (11KB)
	RecommendedDatagramSize = 11264
)

// Default UDP port for datagram operations
const (
	DefaultDatagramUDPPort = 7655
)
