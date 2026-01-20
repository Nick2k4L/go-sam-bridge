// Package handler implements SAM command handlers per SAMv3.md specification.
package handler

import (
	"strings"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

// PingHandler handles PING/PONG commands per SAM 3.2.
// PING echoes arbitrary text back as PONG with the same text.
// No session is required for PING/PONG.
//
// Per SAMv3.md:
//
//	-> PING[ arbitrary text]
//	<- PONG[ arbitrary text]
type PingHandler struct{}

// NewPingHandler creates a new PING handler.
func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

// Handle processes a PING command and returns a PONG response.
// Per SAM 3.2, PING/PONG is used for keepalive and echoes any text.
func (h *PingHandler) Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
	// Extract any text after PING
	// The command.Raw contains the full line, e.g., "PING hello world"
	text := extractPingText(cmd.Raw)

	// Build PONG response with the same text
	return buildPongResponse(text), nil
}

// extractPingText extracts the text portion after "PING " from the raw command.
// Returns empty string if no text follows PING.
func extractPingText(raw string) string {
	// Handle case-insensitive PING prefix
	upper := strings.ToUpper(raw)
	idx := strings.Index(upper, "PING")
	if idx == -1 {
		return ""
	}

	// Skip "PING" and optional space
	rest := raw[idx+4:]
	if len(rest) > 0 && rest[0] == ' ' {
		return rest[1:]
	}
	return rest
}

// buildPongResponse creates a PONG response with optional text.
// Per SAM 3.2, PONG mirrors the format of PING.
func buildPongResponse(text string) *protocol.Response {
	resp := protocol.NewResponse("PONG")
	if text != "" {
		// PONG with arbitrary text - append directly as the response
		// We use a custom format since PONG doesn't use standard KEY=VALUE
		resp.Options = append(resp.Options, text)
	}
	return resp
}

// RegisterPingHandler registers the PING handler with a router.
// PING is a standalone command with no action.
func RegisterPingHandler(router *Router) {
	handler := NewPingHandler()
	router.Register("PING", handler)
}
