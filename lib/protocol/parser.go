// Package protocol provides SAM protocol parsing functionality.
package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/go-i2p/go-sam-bridge/lib/util"
)

// Parser handles parsing of SAM protocol commands.
// The SAM protocol uses text-based commands with key=value options.
//
// Design: Uses bufio.Scanner for line-based reading.
// Supports UTF-8 per SAM 3.2+ and quoted value escaping.
type Parser struct {
	scanner *bufio.Scanner
}

// NewParser creates a new SAM command parser.
func NewParser(r io.Reader) *Parser {
	scanner := bufio.NewScanner(r)
	// SAM protocol lines shouldn't exceed 64KB
	scanner.Buffer(make([]byte, 4096), 65536)
	return &Parser{
		scanner: scanner,
	}
}

// ParseCommand reads and parses the next SAM command.
// Returns io.EOF when no more commands are available.
//
// Command format: VERB [ACTION] [KEY=VALUE ...]
// Example: SESSION CREATE STYLE=STREAM ID=mySession
func (p *Parser) ParseCommand() (*Command, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := strings.TrimSpace(p.scanner.Text())
	if line == "" {
		return nil, util.NewProtocolError("parse", "empty command line")
	}

	cmd := &Command{
		RawLine: line,
		Options: make(map[string]string),
	}

	// Split into tokens (space-separated, respecting quotes)
	tokens, err := tokenize(line)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, util.NewProtocolError("parse", "no tokens in command")
	}

	// First token is always the verb
	cmd.Verb = strings.ToUpper(tokens[0])
	pos := 1

	// Check if second token is an action (doesn't contain '=')
	if len(tokens) > 1 && !strings.Contains(tokens[1], "=") {
		cmd.Action = strings.ToUpper(tokens[1])
		pos = 2
	}

	// Remaining tokens are key=value options
	for i := pos; i < len(tokens); i++ {
		key, value, err := parseOption(tokens[i])
		if err != nil {
			return nil, util.NewProtocolError(cmd.FullCommand(), err.Error())
		}
		cmd.Options[key] = value
	}

	return cmd, nil
}

// tokenize splits a command line into tokens, respecting quoted strings.
// Quoted strings can contain spaces and use backslash escaping.
//
// Examples:
//
//	"KEY=VALUE" -> ["KEY=VALUE"]
//	"KEY=\"quoted value\"" -> ["KEY=quoted value"]
//	"KEY=value\\ with\\ spaces" -> ["KEY=value with spaces"]
func tokenize(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuote := false
	escaped := false

	for _, ch := range line {
		if escaped {
			// Handle escape sequences
			current.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inQuote = !inQuote
			continue
		}

		if ch == ' ' && !inQuote {
			// End of token
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(ch)
	}

	if inQuote {
		return nil, fmt.Errorf("unclosed quote in command")
	}

	if escaped {
		return nil, fmt.Errorf("trailing backslash in command")
	}

	// Add final token
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// parseOption parses a KEY=VALUE option.
// Keys are case-sensitive per SAM specification.
// Values can be empty (KEY=).
func parseOption(token string) (string, string, error) {
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid option format: %s (expected KEY=VALUE)", token)
	}

	key := parts[0]
	value := parts[1]

	if key == "" {
		return "", "", fmt.Errorf("empty key in option")
	}

	return key, value, nil
}

// ParseOptions is a utility function to parse a standalone option string.
// Useful for testing or parsing options from other sources.
func ParseOptions(optString string) (map[string]string, error) {
	options := make(map[string]string)

	tokens, err := tokenize(optString)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		key, value, err := parseOption(token)
		if err != nil {
			return nil, err
		}
		options[key] = value
	}

	return options, nil
}
