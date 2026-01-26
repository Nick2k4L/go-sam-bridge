# util
--
    import "github.com/go-i2p/go-sam-bridge/lib/util"

![util.svg](util.svg)

Package util provides common utilities for the SAM bridge implementation. This
includes custom error types, validation helpers, and logging utilities.

## Usage

```go
var (
	// ErrDuplicateID indicates a session ID already exists.
	// Maps to RESULT=DUPLICATED_ID per SAM spec.
	ErrDuplicateID = errors.New("duplicated session ID")

	// ErrDuplicateDest indicates an I2P destination is already in use.
	// Maps to RESULT=DUPLICATED_DEST per SAM spec.
	ErrDuplicateDest = errors.New("duplicated destination")

	// ErrSessionNotFound indicates the requested session does not exist.
	// Maps to RESULT=INVALID_ID per SAM spec.
	ErrSessionNotFound = errors.New("session not found")

	// ErrInvalidKey indicates the destination key is malformed or invalid.
	// Maps to RESULT=INVALID_KEY per SAM spec.
	ErrInvalidKey = errors.New("invalid key")

	// ErrTimeout indicates an operation timed out.
	// Maps to RESULT=TIMEOUT per SAM spec.
	ErrTimeout = errors.New("timeout")

	// ErrCantReachPeer indicates the remote peer is unreachable.
	// Maps to RESULT=CANT_REACH_PEER per SAM spec.
	ErrCantReachPeer = errors.New("can't reach peer")

	// ErrPeerNotFound indicates the remote peer's destination was not found.
	// Maps to RESULT=PEER_NOT_FOUND per SAM spec.
	ErrPeerNotFound = errors.New("peer not found")

	// ErrLeasesetNotFound indicates the leaseset could not be found.
	// Maps to RESULT=LEASESET_NOT_FOUND per SAM spec.
	ErrLeasesetNotFound = errors.New("leaseset not found")

	// ErrKeyNotFound indicates a name lookup failed.
	// Maps to RESULT=KEY_NOT_FOUND per SAM spec.
	ErrKeyNotFound = errors.New("key not found")

	// ErrNoVersion indicates version negotiation failed.
	// Maps to RESULT=NOVERSION per SAM spec.
	ErrNoVersion = errors.New("no compatible version")

	// ErrAuthRequired indicates authentication is required.
	ErrAuthRequired = errors.New("authentication required")

	// ErrAuthFailed indicates authentication failed.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrSessionClosed indicates the session has been closed.
	ErrSessionClosed = errors.New("session closed")

	// ErrNotImplemented indicates a feature is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrSilentClose indicates the connection should be closed silently
	// without sending any response. Used when SILENT=true and an operation fails.
	// Per SAMv3.md: "If SILENT=true is passed, the SAM bridge won't issue any
	// other message on the socket. If the connection fails, the socket will be closed."
	ErrSilentClose = errors.New("silent close requested")
)
```
Sentinel errors for SAM protocol operations. These map directly to SAM protocol
RESULT codes per SAMv3.md specification.

#### func  IsPermanent

```go
func IsPermanent(err error) bool
```
IsPermanent returns true if the error represents a permanent failure that will
not succeed on retry (e.g., invalid key, auth failed).

#### func  IsRetryable

```go
func IsRetryable(err error) bool
```
IsRetryable returns true if the error represents a condition that may succeed if
retried (e.g., timeout, temporary network issues).

#### func  IsSilentClose

```go
func IsSilentClose(err error) bool
```
IsSilentClose returns true if the error indicates the connection should be
closed without sending a response (SILENT=true behavior).

#### func  ToResultCode

```go
func ToResultCode(err error) string
```
ToResultCode converts a sentinel error to a SAM protocol RESULT code. Returns
"I2P_ERROR" for unknown errors.

#### type ConnectionError

```go
type ConnectionError struct {
	RemoteAddr string // Remote address of the connection
	Operation  string // The operation being performed
	Err        error  // The underlying error
}
```

ConnectionError wraps an error with connection context. Use this when an error
occurs at the connection level.

#### func  NewConnectionError

```go
func NewConnectionError(remoteAddr, operation string, err error) *ConnectionError
```
NewConnectionError creates a new ConnectionError with context.

#### func (*ConnectionError) Error

```go
func (e *ConnectionError) Error() string
```
Error implements the error interface.

#### func (*ConnectionError) Unwrap

```go
func (e *ConnectionError) Unwrap() error
```
Unwrap returns the underlying error for errors.Is and errors.As support.

#### type ProtocolError

```go
type ProtocolError struct {
	Verb    string // The command verb (e.g., "SESSION", "STREAM")
	Action  string // The command action (e.g., "CREATE", "CONNECT")
	Message string // Human-readable error message
	Err     error  // The underlying error (optional)
}
```

ProtocolError wraps an error with SAM protocol command context. Use this when an
error occurs during command parsing or handling.

#### func  NewProtocolError

```go
func NewProtocolError(verb, action, message string) *ProtocolError
```
NewProtocolError creates a new ProtocolError with context.

#### func  NewProtocolErrorWithCause

```go
func NewProtocolErrorWithCause(verb, action, message string, err error) *ProtocolError
```
NewProtocolErrorWithCause creates a new ProtocolError with an underlying cause.

#### func (*ProtocolError) Error

```go
func (e *ProtocolError) Error() string
```
Error implements the error interface.

#### func (*ProtocolError) Unwrap

```go
func (e *ProtocolError) Unwrap() error
```
Unwrap returns the underlying error for errors.Is and errors.As support.

#### type SessionError

```go
type SessionError struct {
	SessionID string // The session ID where the error occurred
	Operation string // The operation being performed (e.g., "connect", "accept")
	Err       error  // The underlying error
}
```

SessionError wraps an error with session context. Use this when an error occurs
during session operations.

#### func  NewSessionError

```go
func NewSessionError(sessionID, operation string, err error) *SessionError
```
NewSessionError creates a new SessionError with context.

#### func (*SessionError) Error

```go
func (e *SessionError) Error() string
```
Error implements the error interface.

#### func (*SessionError) Unwrap

```go
func (e *SessionError) Unwrap() error
```
Unwrap returns the underlying error for errors.Is and errors.As support.

#### type SilentCloseError

```go
type SilentCloseError struct {
	Operation string // The operation that failed (e.g., "connect", "accept")
	Err       error  // The underlying error
}
```

SilentCloseError wraps an error that should cause the connection to be closed
silently without sending any response. This is used when SILENT=true is set and
an operation fails. Per SAMv3.md: "If SILENT=true is passed, the SAM bridge
won't issue any other message on the socket. If the connection fails, the socket
will be closed."

#### func  NewSilentCloseError

```go
func NewSilentCloseError(operation string, err error) *SilentCloseError
```
NewSilentCloseError creates a new SilentCloseError.

#### func (*SilentCloseError) Error

```go
func (e *SilentCloseError) Error() string
```
Error implements the error interface.

#### func (*SilentCloseError) Unwrap

```go
func (e *SilentCloseError) Unwrap() error
```
Unwrap returns the underlying error for errors.Is and errors.As support.



util 

github.com/go-i2p/go-sam-bridge/lib/util

[go-i2p template file](/template.md)
