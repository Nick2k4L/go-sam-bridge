package protocol

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Parser errors
var (
	ErrEmptyCommand      = errors.New("empty command")
	ErrInvalidUTF8       = errors.New("command contains invalid UTF-8")
	ErrUnterminatedQuote = errors.New("unterminated quoted value")
	ErrInvalidEscape     = errors.New("invalid escape sequence")
)

// Parser tokenizes SAM protocol commands.
// Per SAMv3.md, commands follow the format:
//
//	VERB [ACTION] [KEY=VALUE]...
//
// Parser handles UTF-8 encoding (SAM 3.2+), quoted values with escapes,
// and empty option values.
type Parser struct {
	// CaseInsensitive enables case-insensitive verb/action matching.
	// Per SAM spec, this is recommended but not required.
	CaseInsensitive bool
}

// NewParser creates a new parser with default settings.
// Case-insensitive matching is enabled by default per SAM spec recommendation.
func NewParser() *Parser {
	return &Parser{
		CaseInsensitive: true,
	}
}

// Parse parses a SAM command line into a Command struct.
// The input should be a single line without the trailing newline.
func (p *Parser) Parse(line string) (*Command, error) {
	// Trim trailing newline/carriage return if present
	line = strings.TrimRight(line, "\r\n")

	// Store raw command for debugging
	raw := line

	// Validate UTF-8 encoding (SAM 3.2 requirement)
	if !utf8.ValidString(line) {
		return nil, ErrInvalidUTF8
	}

	// Tokenize the command
	tokens, err := p.tokenize(line)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, ErrEmptyCommand
	}

	cmd := &Command{
		Options: make(map[string]string),
		Raw:     raw,
	}

	// First token is always the verb
	verb := tokens[0]
	if p.CaseInsensitive {
		verb = strings.ToUpper(verb)
	}
	cmd.Verb = verb

	// Process remaining tokens
	tokenIdx := 1

	// Check if second token is an action (doesn't contain '=')
	if tokenIdx < len(tokens) && !strings.Contains(tokens[tokenIdx], "=") {
		action := tokens[tokenIdx]
		// Check if it looks like a known action or just a standalone word
		// (e.g., PING has optional text after it, not an action)
		if p.isAction(verb, action) {
			if p.CaseInsensitive {
				action = strings.ToUpper(action)
			}
			cmd.Action = action
			tokenIdx++
		}
	}

	// Remaining tokens are key=value pairs
	for ; tokenIdx < len(tokens); tokenIdx++ {
		token := tokens[tokenIdx]
		key, value := p.parseKeyValue(token)
		if key != "" {
			cmd.Options[key] = value
		}
	}

	return cmd, nil
}

// tokenize splits a command line into tokens, handling quoted values.
func (p *Parser) tokenize(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuote := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if escaped {
			// Handle escape sequences per SAM 3.2
			switch ch {
			case '"', '\\':
				current.WriteByte(ch)
			default:
				// Invalid escape - include backslash and character
				current.WriteByte('\\')
				current.WriteByte(ch)
			}
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			if inQuote {
				escaped = true
			} else {
				current.WriteByte(ch)
			}

		case '"':
			inQuote = !inQuote
			// Include quote in token to mark it was quoted
			current.WriteByte(ch)

		case ' ', '\t':
			if inQuote {
				current.WriteByte(ch)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			// Multiple spaces are allowed per SAM 3.2

		default:
			current.WriteByte(ch)
		}
	}

	if inQuote {
		return nil, ErrUnterminatedQuote
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// parseKeyValue parses a token as a key=value pair.
// Handles empty values per SAM 3.2 (KEY, KEY=, KEY="").
func (p *Parser) parseKeyValue(token string) (key, value string) {
	// Find the first '=' that's not inside quotes
	eqIdx := -1
	inQuote := false
	for i := 0; i < len(token); i++ {
		if token[i] == '"' {
			inQuote = !inQuote
		} else if token[i] == '=' && !inQuote {
			eqIdx = i
			break
		}
	}

	if eqIdx < 0 {
		// No '=' found - empty value (KEY format per SAM 3.2)
		return token, ""
	}

	key = token[:eqIdx]
	value = token[eqIdx+1:]

	// Strip quotes from value if present
	value = stripQuotes(value)

	return key, value
}

// stripQuotes removes surrounding quotes and unescapes the value.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		// Unescape
		s = strings.ReplaceAll(s, "\\\"", "\"")
		s = strings.ReplaceAll(s, "\\\\", "\\")
	}
	return s
}

// isAction determines if a token should be treated as an action.
// Per SAM spec, some commands like PING don't have actions.
func (p *Parser) isAction(verb, token string) bool {
	// Normalize for comparison
	v := strings.ToUpper(verb)
	t := strings.ToUpper(token)

	// Known verb+action combinations
	switch v {
	case VerbHello:
		return t == ActionVersion
	case VerbSession:
		return t == ActionCreate || t == ActionAdd || t == ActionRemove
	case VerbStream:
		return t == ActionConnect || t == ActionAccept || t == ActionForward
	case VerbDatagram, VerbRaw:
		return t == ActionSend || t == ActionReceived
	case VerbDest:
		return t == ActionGenerate || t == ActionReply
	case VerbNaming:
		return t == ActionLookup
	case VerbAuth:
		return t == ActionEnable || t == ActionDisable || t == ActionAdd || t == ActionRemove
	case VerbPing, VerbPong, VerbQuit, VerbStop, VerbExit, VerbHelp:
		// These commands don't have actions
		return false
	default:
		// For unknown verbs, treat it as action if it doesn't contain '='
		return !strings.Contains(token, "=")
	}
}

// ParseLine is a convenience function that parses a line using default settings.
func ParseLine(line string) (*Command, error) {
	return NewParser().Parse(line)
}

// MustParse parses a line and panics on error. For testing only.
func MustParse(line string) *Command {
	cmd, err := ParseLine(line)
	if err != nil {
		panic(fmt.Sprintf("failed to parse command: %v", err))
	}
	return cmd
}
