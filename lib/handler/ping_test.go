package handler

import (
	"testing"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

func TestPingHandler_Handle(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantText string
	}{
		{
			name:     "PING with no text",
			raw:      "PING",
			wantText: "",
		},
		{
			name:     "PING with simple text",
			raw:      "PING hello",
			wantText: "hello",
		},
		{
			name:     "PING with multiple words",
			raw:      "PING hello world test",
			wantText: "hello world test",
		},
		{
			name:     "PING with numbers",
			raw:      "PING 12345",
			wantText: "12345",
		},
		{
			name:     "PING lowercase",
			raw:      "ping keepalive",
			wantText: "keepalive",
		},
		{
			name:     "PING mixed case",
			raw:      "Ping test",
			wantText: "test",
		},
		{
			name:     "PING with special characters",
			raw:      "PING test=value foo:bar",
			wantText: "test=value foo:bar",
		},
		{
			name:     "PING with unicode",
			raw:      "PING こんにちは",
			wantText: "こんにちは",
		},
		{
			name:     "PING with leading/trailing spaces in text",
			raw:      "PING  multiple  spaces",
			wantText: " multiple  spaces",
		},
	}

	handler := NewPingHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &protocol.Command{
				Verb: "PING",
				Raw:  tt.raw,
			}

			resp, err := handler.Handle(nil, cmd)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Handle() returned nil response")
			}

			// Check verb is PONG
			if resp.Verb != "PONG" {
				t.Errorf("response Verb = %q, want %q", resp.Verb, "PONG")
			}

			// Check text matches
			respStr := resp.String()
			if tt.wantText == "" {
				// Should be just "PONG\n"
				if respStr != "PONG\n" {
					t.Errorf("response = %q, want %q", respStr, "PONG\n")
				}
			} else {
				// Should be "PONG text\n"
				want := "PONG " + tt.wantText + "\n"
				if respStr != want {
					t.Errorf("response = %q, want %q", respStr, want)
				}
			}
		})
	}
}

func TestExtractPingText(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"no text", "PING", ""},
		{"simple text", "PING hello", "hello"},
		{"multiple words", "PING hello world", "hello world"},
		{"lowercase ping", "ping test", "test"},
		{"mixed case", "PiNg test", "test"},
		{"no space after ping", "PINGtest", "test"},
		{"empty string", "", ""},
		{"no ping", "HELLO", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPingText(tt.raw)
			if got != tt.want {
				t.Errorf("extractPingText(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestBuildPongResponse(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"empty text", "", "PONG\n"},
		{"simple text", "hello", "PONG hello\n"},
		{"multiple words", "hello world", "PONG hello world\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := buildPongResponse(tt.text)
			got := resp.String()
			if got != tt.want {
				t.Errorf("buildPongResponse(%q).String() = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// TestPongWithSpecialCharacters verifies PONG responses with special characters
// are formatted correctly per SAM 3.2 specification. Per SAMv3.md, PONG should
// echo the arbitrary text exactly as received without additional quoting.
func TestPongWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		// Per SAM spec, PONG echoes arbitrary text without quoting
		{"text with equals", "foo=bar", "PONG foo=bar\n"},
		{"text with colon", "host:port", "PONG host:port\n"},
		{"text with quotes", `"quoted"`, `PONG "quoted"` + "\n"},
		{"text with backslash", `path\to\file`, `PONG path\to\file` + "\n"},
		{"text with tabs", "tab\ttab", "PONG tab\ttab\n"},
		{"complex key=value", "KEY=value with spaces", "PONG KEY=value with spaces\n"},
		{"multiple special chars", `a="b" c="d"`, `PONG a="b" c="d"` + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := buildPongResponse(tt.text)
			got := resp.String()
			if got != tt.want {
				t.Errorf("buildPongResponse(%q).String() = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestRegisterPingHandler(t *testing.T) {
	router := NewRouter()
	RegisterPingHandler(router)

	// Verify handler is registered
	cmd := &protocol.Command{Verb: "PING"}
	h := router.Route(cmd)
	if h == nil {
		t.Error("PING handler not registered")
	}
}

func TestPingHandler_NoSessionRequired(t *testing.T) {
	// PING should work even with nil context
	handler := NewPingHandler()
	cmd := &protocol.Command{
		Verb: "PING",
		Raw:  "PING test",
	}

	resp, err := handler.Handle(nil, cmd)
	if err != nil {
		t.Fatalf("Handle() with nil context error = %v", err)
	}
	if resp == nil {
		t.Fatal("Handle() returned nil response")
	}
	if resp.Verb != "PONG" {
		t.Errorf("response Verb = %q, want %q", resp.Verb, "PONG")
	}
}
