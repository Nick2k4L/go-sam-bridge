// Package handler implements SAM command handlers per SAMv3.md specification.
package handler

import (
	"strings"

	"github.com/go-i2p/go-sam-bridge/lib/destination"
	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

// NamingHandler handles NAMING LOOKUP commands per SAM 3.0-3.3.
// Resolves I2P hostnames, .b32.i2p addresses, and special names like ME.
type NamingHandler struct {
	destManager destination.Manager
}

// NewNamingHandler creates a new NAMING handler with the given destination manager.
func NewNamingHandler(destManager destination.Manager) *NamingHandler {
	return &NamingHandler{destManager: destManager}
}

// Handle processes a NAMING LOOKUP command.
// Per SAMv3.md, NAMING LOOKUP resolves names to destinations.
//
// Request: NAMING LOOKUP NAME=$name [OPTIONS=true]
// Response: NAMING REPLY RESULT=OK NAME=$name VALUE=$destination
//
//	NAMING REPLY RESULT=KEY_NOT_FOUND NAME=$name
//	NAMING REPLY RESULT=INVALID_KEY NAME=$name MESSAGE="..."
func (h *NamingHandler) Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
	name := cmd.Get("NAME")
	if name == "" {
		return namingInvalidKey("", "missing NAME parameter"), nil
	}

	// Special case: NAME=ME returns session destination
	if name == "ME" {
		return h.handleNameMe(ctx, name)
	}

	// Validate name format
	if !isValidName(name) {
		return namingInvalidKey(name, "invalid name format"), nil
	}

	// Try to resolve the name
	dest, err := h.resolveName(name)
	if err != nil {
		return namingKeyNotFound(name), nil
	}

	return namingOK(name, dest), nil
}

// handleNameMe returns the destination of the current session.
func (h *NamingHandler) handleNameMe(ctx *Context, name string) (*protocol.Response, error) {
	if ctx.Session == nil {
		return namingInvalidKey(name, "no session bound"), nil
	}

	dest := ctx.Session.Destination()
	if dest == nil {
		return namingInvalidKey(name, "session has no destination"), nil
	}

	// Return the public key as base64
	return namingOK(name, string(dest.PublicKey)), nil
}

// resolveName attempts to resolve a name to a destination.
// Supports .i2p hostnames and .b32.i2p addresses.
func (h *NamingHandler) resolveName(name string) (string, error) {
	// Check for .b32.i2p address
	if isB32Address(name) {
		return h.resolveB32(name)
	}

	// Check for .i2p hostname
	if isI2PHostname(name) {
		return h.resolveHostname(name)
	}

	// Check if it's already a Base64 destination
	if isBase64Destination(name) {
		return name, nil
	}

	return "", &namingErr{msg: "unknown name format"}
}

// resolveB32 resolves a .b32.i2p address.
// TODO: This requires network lookup via I2CP.
func (h *NamingHandler) resolveB32(name string) (string, error) {
	// For now, return not found - actual lookup requires I2CP integration
	return "", &namingErr{msg: "b32 lookup not implemented"}
}

// resolveHostname resolves an .i2p hostname.
// TODO: This requires addressbook lookup or network query.
func (h *NamingHandler) resolveHostname(name string) (string, error) {
	// For now, return not found - actual lookup requires addressbook
	return "", &namingErr{msg: "hostname lookup not implemented"}
}

// isValidName checks if a name is valid for lookup.
func isValidName(name string) bool {
	if name == "" {
		return false
	}
	// Check for obviously invalid characters
	if strings.ContainsAny(name, "\n\r\t") {
		return false
	}
	return true
}

// isB32Address checks if the name is a .b32.i2p address.
func isB32Address(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".b32.i2p")
}

// isI2PHostname checks if the name is an .i2p hostname (not b32).
func isI2PHostname(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".i2p") && !strings.HasSuffix(lower, ".b32.i2p")
}

// isBase64Destination checks if the name looks like a base64 destination.
// Destinations are 516+ base64 characters.
func isBase64Destination(name string) bool {
	if len(name) < 516 {
		return false
	}
	// Check for valid base64 characters (I2P alphabet)
	for _, c := range name {
		if !isBase64Char(c) {
			return false
		}
	}
	return true
}

// isBase64Char checks if a rune is a valid I2P Base64 character.
func isBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '~'
}

// namingOK returns a successful NAMING REPLY response.
func namingOK(name, destination string) *protocol.Response {
	return protocol.NewResponse(protocol.VerbNaming).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultOK).
		WithOption("NAME", name).
		WithOption("VALUE", destination)
}

// namingKeyNotFound returns a KEY_NOT_FOUND response.
func namingKeyNotFound(name string) *protocol.Response {
	return protocol.NewResponse(protocol.VerbNaming).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultKeyNotFound).
		WithOption("NAME", name)
}

// namingInvalidKey returns an INVALID_KEY response.
func namingInvalidKey(name, msg string) *protocol.Response {
	resp := protocol.NewResponse(protocol.VerbNaming).
		WithAction(protocol.ActionReply).
		WithResult(protocol.ResultInvalidKey)
	if name != "" {
		resp = resp.WithOption("NAME", name)
	}
	if msg != "" {
		resp = resp.WithMessage(msg)
	}
	return resp
}

// namingErr is an error type for naming lookup errors.
type namingErr struct {
	msg string
}

func (e *namingErr) Error() string {
	return e.msg
}

// Ensure NamingHandler implements Handler interface
var _ Handler = (*NamingHandler)(nil)
