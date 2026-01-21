package protocol

// Command represents a parsed SAM protocol command.
// Per SAMv3.md, commands follow the format:
//
//	VERB [ACTION] [KEY=VALUE]...
//
// All SAM messages are sent on a single line terminated by newline.
// For commands with binary payloads (e.g., RAW SEND, DATAGRAM SEND),
// the payload follows the command line.
type Command struct {
	// Verb is the primary command (e.g., "HELLO", "SESSION", "STREAM").
	Verb string

	// Action is the secondary command (e.g., "VERSION", "CREATE", "CONNECT").
	// May be empty for commands like PING per SAM 3.2.
	Action string

	// Options contains key-value pairs from the command.
	// Keys are case-sensitive per SAM specification.
	// Empty values are allowed per SAM 3.2 (KEY, KEY=, KEY="").
	Options map[string]string

	// Payload contains binary data following the command line.
	// Used by RAW SEND, DATAGRAM SEND commands per SAMv3.md.
	// The payload size is specified in the SIZE option.
	Payload []byte

	// Raw is the original command line for debugging and logging.
	Raw string
}

// NewCommand creates a new Command with initialized Options map.
func NewCommand(verb, action string) *Command {
	return &Command{
		Verb:    verb,
		Action:  action,
		Options: make(map[string]string),
	}
}

// Get returns an option value, or empty string if not present.
// Use Has() to distinguish between missing keys and empty values.
func (c *Command) Get(key string) string {
	if c.Options == nil {
		return ""
	}
	return c.Options[key]
}

// GetOrDefault returns an option value, or the default if not present.
// Note: If the key is present but empty, the empty value is returned,
// not the default. Use Has() to check for key presence.
func (c *Command) GetOrDefault(key, defaultVal string) string {
	if c.Options == nil {
		return defaultVal
	}
	if v, ok := c.Options[key]; ok {
		return v
	}
	return defaultVal
}

// Has returns true if the option key is present (even if empty).
// Per SAM 3.2, empty option values such as KEY, KEY=, or KEY=""
// may be allowed, implementation dependent.
func (c *Command) Has(key string) bool {
	if c.Options == nil {
		return false
	}
	_, ok := c.Options[key]
	return ok
}

// Set adds or updates an option value.
func (c *Command) Set(key, value string) {
	if c.Options == nil {
		c.Options = make(map[string]string)
	}
	c.Options[key] = value
}
