package handler

import (
	"strings"
	"sync"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

// Router dispatches SAM commands to appropriate handlers.
// Per SAMv3.md, it is recommended that servers map commands to upper case
// for ease in testing via telnet.
type Router struct {
	mu       sync.RWMutex
	handlers map[string]Handler

	// CaseInsensitive enables case-insensitive verb/action matching.
	// Recommended per SAM 3.2 specification.
	CaseInsensitive bool

	// UnknownHandler is called when no handler matches the command.
	// If nil, returns I2P_ERROR with "unknown command" message.
	UnknownHandler Handler
}

// NewRouter creates a new command router with case-insensitive matching enabled.
func NewRouter() *Router {
	return &Router{
		handlers:        make(map[string]Handler),
		CaseInsensitive: true,
	}
}

// Register adds a handler for a command key.
// The key format is "VERB" or "VERB ACTION" (e.g., "HELLO VERSION").
// If CaseInsensitive is true, the key is normalized to upper case.
func (r *Router) Register(key string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.CaseInsensitive {
		key = strings.ToUpper(key)
	}
	r.handlers[key] = handler
}

// RegisterFunc is a convenience method to register a HandlerFunc.
func (r *Router) RegisterFunc(key string, fn HandlerFunc) {
	r.Register(key, fn)
}

// Route returns the handler for the given command.
// Matching order:
// 1. "VERB ACTION" (exact match)
// 2. "VERB" (verb-only match for commands like PING, QUIT)
// 3. UnknownHandler (if set)
// 4. nil (no handler found)
func (r *Router) Route(cmd *protocol.Command) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	verb := cmd.Verb
	action := cmd.Action

	if r.CaseInsensitive {
		verb = strings.ToUpper(verb)
		action = strings.ToUpper(action)
	}

	// Try "VERB ACTION" first
	if action != "" {
		key := verb + " " + action
		if h, ok := r.handlers[key]; ok {
			return h
		}
	}

	// Try "VERB" only
	if h, ok := r.handlers[verb]; ok {
		return h
	}

	// Fall back to unknown handler
	return r.UnknownHandler
}

// Handle dispatches the command to the appropriate handler.
// If no handler is found and UnknownHandler is nil, returns an I2P_ERROR response.
func (r *Router) Handle(ctx *Context, cmd *protocol.Command) (*protocol.Response, error) {
	handler := r.Route(cmd)
	if handler == nil {
		return r.unknownCommandResponse(cmd), nil
	}
	return handler.Handle(ctx, cmd)
}

// unknownCommandResponse builds an error response for unknown commands.
func (r *Router) unknownCommandResponse(cmd *protocol.Command) *protocol.Response {
	// Use the verb from the command if available, otherwise use a generic response
	verb := cmd.Verb
	if verb == "" {
		verb = "ERROR"
	}

	return protocol.NewResponse(verb).
		WithAction(protocol.ActionStatus).
		WithResult(protocol.ResultI2PError).
		WithMessage("unknown command")
}

// HasHandler returns true if a handler is registered for the given key.
func (r *Router) HasHandler(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.CaseInsensitive {
		key = strings.ToUpper(key)
	}
	_, ok := r.handlers[key]
	return ok
}

// Keys returns all registered handler keys.
func (r *Router) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.handlers))
	for k := range r.handlers {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of registered handlers.
func (r *Router) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}
