// Package session provides SAM session management and implementations.
package session

import (
	"net"

	i2cp "github.com/go-i2p/go-i2cp"
)

// Session represents the base interface for all SAM session types.
// This interface defines common lifecycle and communication methods
// shared across STREAM, DATAGRAM, RAW, and PRIMARY session styles.
type Session interface {
	// ID returns the unique identifier for this session
	ID() string

	// Destination returns the I2P destination for this session
	Destination() *Destination

	// Close terminates the session and releases resources
	Close() error

	// IsClosed returns true if the session has been closed
	IsClosed() bool

	// I2CPSession returns the underlying I2CP session for tunnel communication
	I2CPSession() *i2cp.Session

	// Config returns the session configuration
	Config() *SessionConfig

	// Style returns the session style (STREAM, DATAGRAM, RAW, etc.)
	Style() SessionStyle

	// Status returns the current session state
	Status() SessionStatus
}

// StreamSession extends Session for virtual streaming connections.
// Implements SAM 3.0/3.1 STREAM session style with CONNECT, ACCEPT, and FORWARD.
type StreamSession interface {
	Session

	// Connect establishes a streaming connection to the destination
	Connect(dest *Destination, opts ConnectOptions) (net.Conn, error)

	// Accept waits for an incoming streaming connection
	Accept() (net.Conn, *Destination, error)

	// Forward starts forwarding incoming connections to host:port
	Forward(host string, port int, opts ForwardOptions) error

	// StopForward stops the forwarding behavior
	StopForward() error
}

// DatagramSession extends Session for repliable authenticated datagrams.
// Implements SAM 3.0/3.1 DATAGRAM session style.
type DatagramSession interface {
	Session

	// Send transmits a datagram to the destination
	Send(dest *Destination, data []byte, opts DatagramOptions) error

	// Receive waits for an incoming datagram
	Receive() (data []byte, from *Destination, metadata DatagramMetadata, err error)

	// StartForwarding forwards received datagrams to host:port via UDP
	StartForwarding(host string, port int) error

	// StopForwarding stops UDP forwarding
	StopForwarding() error
}

// RawSession extends Session for anonymous datagrams without authentication.
// Implements SAM 3.0/3.1 RAW session style.
type RawSession interface {
	Session

	// Send transmits a raw datagram to the destination
	Send(dest *Destination, data []byte, opts RawOptions) error

	// Receive waits for an incoming raw datagram
	Receive() (data []byte, metadata RawMetadata, err error)

	// StartForwarding forwards raw datagrams to host:port via UDP
	StartForwarding(host string, port int, headerMode bool) error

	// StopForwarding stops UDP forwarding
	StopForwarding() error
}

// PrimarySession extends Session for multiplexed subsessions (SAM 3.3).
// A PRIMARY session can host multiple subsessions of different styles
// sharing the same I2P destination.
type PrimarySession interface {
	Session

	// AddSubsession creates a new subsession with the given configuration
	AddSubsession(config SubsessionConfig) (Session, error)

	// RemoveSubsession deletes a subsession by ID
	RemoveSubsession(id string) error

	// GetSubsession retrieves a subsession by ID
	GetSubsession(id string) (Session, bool)

	// ListSubsessions returns all subsession IDs
	ListSubsessions() []string

	// Route determines which subsession should handle incoming traffic
	Route(protocol int, toPort int, fromPort int) (Session, error)
}

// SessionStyle represents the type of SAM session
type SessionStyle int

const (
	// StyleStream is for reliable ordered streaming connections
	StyleStream SessionStyle = iota
	// StyleDatagram is for repliable authenticated datagrams
	StyleDatagram
	// StyleRaw is for anonymous datagrams without authentication
	StyleRaw
	// StyleDatagram2 is enhanced datagram format (SAM 3.2+)
	StyleDatagram2
	// StyleDatagram3 is repliable but unauthenticated datagrams (SAM 3.2+)
	StyleDatagram3
	// StylePrimary is for multiplexed subsessions (SAM 3.3)
	StylePrimary
)

// String returns the SAM protocol string for the session style
func (s SessionStyle) String() string {
	switch s {
	case StyleStream:
		return "STREAM"
	case StyleDatagram:
		return "DATAGRAM"
	case StyleRaw:
		return "RAW"
	case StyleDatagram2:
		return "DATAGRAM2"
	case StyleDatagram3:
		return "DATAGRAM3"
	case StylePrimary:
		return "PRIMARY"
	default:
		return "UNKNOWN"
	}
}

// SessionStatus represents the lifecycle state of a session
type SessionStatus int

const (
	// StatusCreating indicates the session is being initialized
	StatusCreating SessionStatus = iota
	// StatusReady indicates the session is ready for operations
	StatusReady
	// StatusClosing indicates the session is shutting down
	StatusClosing
	// StatusClosed indicates the session is fully closed
	StatusClosed
	// StatusError indicates the session encountered a fatal error
	StatusError
)

// String returns a human-readable status description
func (s SessionStatus) String() string {
	switch s {
	case StatusCreating:
		return "CREATING"
	case StatusReady:
		return "READY"
	case StatusClosing:
		return "CLOSING"
	case StatusClosed:
		return "CLOSED"
	case StatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
