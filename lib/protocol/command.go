// Package protocol provides SAM protocol command structures.
package protocol

// Command represents a parsed SAM protocol command.
// Commands follow the format: VERB [ARGS...] [KEY=VALUE ...]
//
// Examples:
//
//	HELLO VERSION MIN=3.0 MAX=3.3
//	SESSION CREATE STYLE=STREAM ID=mySession
//	DEST GENERATE SIGNATURE_TYPE=7
type Command struct {
	// Verb is the primary command (e.g., "HELLO", "SESSION")
	Verb string

	// Action is the subcommand (e.g., "CREATE", "LOOKUP")
	// Empty for commands without actions (e.g., "PING")
	Action string

	// Options contains key=value pairs from the command
	// Keys are case-sensitive per SAM spec
	Options map[string]string

	// RawLine is the original unparsed command line
	RawLine string
}

// Get retrieves an option value by key.
// Returns empty string if key doesn't exist.
func (c *Command) Get(key string) string {
	if c.Options == nil {
		return ""
	}
	return c.Options[key]
}

// GetOr retrieves an option value by key with a default fallback.
func (c *Command) GetOr(key, defaultValue string) string {
	if c.Options == nil {
		return defaultValue
	}
	if val, ok := c.Options[key]; ok {
		return val
	}
	return defaultValue
}

// Has checks if an option key exists.
func (c *Command) Has(key string) bool {
	if c.Options == nil {
		return false
	}
	_, ok := c.Options[key]
	return ok
}

// FullCommand returns the verb and action combined.
// For "SESSION CREATE", returns "SESSION CREATE".
// For "PING", returns "PING".
func (c *Command) FullCommand() string {
	if c.Action == "" {
		return c.Verb
	}
	return c.Verb + " " + c.Action
}
