// Package handler implements SAM command handlers per SAMv3.md specification.
package handler

import (
	"strconv"
	"strings"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

// HelloConfig holds configuration for the HELLO handler.
type HelloConfig struct {
	// MinVersion is the minimum SAM version this bridge supports.
	// Default: "3.0"
	MinVersion string

	// MaxVersion is the maximum SAM version this bridge supports.
	// Default: "3.3"
	MaxVersion string

	// RequireAuth indicates if authentication is required.
	// If true, USER and PASSWORD must be provided in HELLO.
	RequireAuth bool

	// AuthFunc validates user/password credentials.
	// Only called if RequireAuth is true.
	// Returns true if credentials are valid.
	AuthFunc func(user, password string) bool
}

// DefaultHelloConfig returns the default HELLO configuration.
func DefaultHelloConfig() HelloConfig {
	return HelloConfig{
		MinVersion:  protocol.SAMVersionMin,
		MaxVersion:  protocol.SAMVersionMax,
		RequireAuth: false,
		AuthFunc:    nil,
	}
}

// HelloHandler handles HELLO VERSION commands per SAM 3.0-3.3.
// Performs version negotiation and optional authentication.
type HelloHandler struct {
	config HelloConfig
}

// NewHelloHandler creates a new HELLO handler with the given configuration.
func NewHelloHandler(config HelloConfig) *HelloHandler {
	return &HelloHandler{config: config}
}

// Handle processes a HELLO VERSION command.
// Per SAMv3.md, HELLO must be the first command on a connection.
//
// Request: HELLO VERSION [MIN=$min] [MAX=$max] [USER="xxx"] [PASSWORD="yyy"]
// Response: HELLO REPLY RESULT=OK VERSION=3.3
//
//	HELLO REPLY RESULT=NOVERSION
//	HELLO REPLY RESULT=I2P_ERROR MESSAGE="..."
func (h *HelloHandler) Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
	// Reject if handshake already complete
	if ctx.HandshakeComplete {
		return helloError("HELLO already completed"), nil
	}

	// Parse client version constraints
	clientMin, clientMax, err := parseVersionRange(cmd)
	if err != nil {
		return helloError(err.Error()), nil
	}

	// Negotiate version
	version, ok := h.negotiateVersion(clientMin, clientMax)
	if !ok {
		return helloNoVersion(), nil
	}

	// Handle authentication if required
	if h.config.RequireAuth {
		if !h.authenticate(cmd) {
			return helloError("Authentication failed"), nil
		}
		ctx.Authenticated = true
	}

	// Update context state
	ctx.Version = version
	ctx.HandshakeComplete = true

	return helloOK(version), nil
}

// parseVersionRange extracts MIN and MAX version from command.
// Returns defaults if not specified (SAM 3.1+ behavior).
func parseVersionRange(cmd *protocol.Command) (min, max string, err error) {
	min = cmd.GetOrDefault("MIN", protocol.SAMVersionMin)
	max = cmd.GetOrDefault("MAX", protocol.SAMVersionMax)

	// Validate version format
	if !isValidVersion(min) {
		return "", "", &versionError{msg: "invalid MIN version format"}
	}
	if !isValidVersion(max) {
		return "", "", &versionError{msg: "invalid MAX version format"}
	}

	// Validate range
	if compareVersions(min, max) > 0 {
		return "", "", &versionError{msg: "MIN version cannot be greater than MAX"}
	}

	return min, max, nil
}

// negotiateVersion finds the highest compatible version.
// Returns empty string if no compatible version exists.
func (h *HelloHandler) negotiateVersion(clientMin, clientMax string) (string, bool) {
	// Server's supported range
	serverMin := h.config.MinVersion
	serverMax := h.config.MaxVersion

	// Find overlap: max of mins and min of maxes
	overlapMin := laterVersion(clientMin, serverMin)
	overlapMax := earlierVersion(clientMax, serverMax)

	// Check if valid overlap exists
	if compareVersions(overlapMin, overlapMax) > 0 {
		return "", false
	}

	// Return highest compatible version (overlapMax)
	return overlapMax, true
}

// authenticate validates USER and PASSWORD credentials.
func (h *HelloHandler) authenticate(cmd *protocol.Command) bool {
	if h.config.AuthFunc == nil {
		return false
	}

	user := cmd.Get("USER")
	password := cmd.Get("PASSWORD")

	// Both must be provided if auth is required
	if user == "" || password == "" {
		return false
	}

	return h.config.AuthFunc(user, password)
}

// helloOK returns a successful HELLO REPLY.
func helloOK(version string) *protocol.Response {
	return protocol.NewResponse(protocol.VerbHello).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultOK).
		WithVersion(version)
}

// helloNoVersion returns a NOVERSION response.
func helloNoVersion() *protocol.Response {
	return protocol.NewResponse(protocol.VerbHello).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultNoVersion)
}

// helloError returns an I2P_ERROR response with a message.
func helloError(msg string) *protocol.Response {
	return protocol.NewResponse(protocol.VerbHello).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultI2PError).
		WithMessage(msg)
}

// versionError represents a version parsing or validation error.
type versionError struct {
	msg string
}

func (e *versionError) Error() string {
	return e.msg
}

// isValidVersion checks if a version string is in format "X.Y".
func isValidVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return false
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

// parseVersion splits a version string into major and minor components.
func parseVersion(v string) (major, minor int) {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return 0, 0
	}
	major, _ = strconv.Atoi(parts[0])
	minor, _ = strconv.Atoi(parts[1])
	return major, minor
}

// compareVersions compares two version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	aMaj, aMin := parseVersion(a)
	bMaj, bMin := parseVersion(b)

	if aMaj != bMaj {
		if aMaj < bMaj {
			return -1
		}
		return 1
	}
	if aMin != bMin {
		if aMin < bMin {
			return -1
		}
		return 1
	}
	return 0
}

// laterVersion returns the later of two versions.
func laterVersion(a, b string) string {
	if compareVersions(a, b) >= 0 {
		return a
	}
	return b
}

// earlierVersion returns the earlier of two versions.
func earlierVersion(a, b string) string {
	if compareVersions(a, b) <= 0 {
		return a
	}
	return b
}
