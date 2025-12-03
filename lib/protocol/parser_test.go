package protocol

import (
	"io"
	"strings"
	"testing"
)

func TestParser_ParseCommand_Simple(t *testing.T) {
	input := "HELLO VERSION\n"
	parser := NewParser(strings.NewReader(input))

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	if cmd.Verb != "HELLO" {
		t.Errorf("Verb = %v, want HELLO", cmd.Verb)
	}

	if cmd.Action != "VERSION" {
		t.Errorf("Action = %v, want VERSION", cmd.Action)
	}
}

func TestParser_ParseCommand_WithOptions(t *testing.T) {
	input := "SESSION CREATE STYLE=STREAM ID=test\n"
	parser := NewParser(strings.NewReader(input))

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	if cmd.Verb != "SESSION" {
		t.Errorf("Verb = %v, want SESSION", cmd.Verb)
	}

	if cmd.Action != "CREATE" {
		t.Errorf("Action = %v, want CREATE", cmd.Action)
	}

	if cmd.Get("STYLE") != "STREAM" {
		t.Errorf("STYLE = %v, want STREAM", cmd.Get("STYLE"))
	}

	if cmd.Get("ID") != "test" {
		t.Errorf("ID = %v, want test", cmd.Get("ID"))
	}
}

func TestParser_ParseCommand_QuotedValue(t *testing.T) {
	input := "NAMING LOOKUP NAME=\"my test.i2p\"\n"
	parser := NewParser(strings.NewReader(input))

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	if cmd.Get("NAME") != "my test.i2p" {
		t.Errorf("NAME = %v, want 'my test.i2p'", cmd.Get("NAME"))
	}
}

func TestParser_ParseCommand_EscapedQuote(t *testing.T) {
	input := "TEST KEY=\"value\\\"quoted\\\"\"\n"
	parser := NewParser(strings.NewReader(input))

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	if cmd.Get("KEY") != "value\"quoted\"" {
		t.Errorf("KEY = %v, want 'value\"quoted\"'", cmd.Get("KEY"))
	}
}

func TestParser_ParseCommand_EmptyValue(t *testing.T) {
	input := "TEST KEY=\n"
	parser := NewParser(strings.NewReader(input))

	cmd, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	if !cmd.Has("KEY") {
		t.Error("Should have KEY option")
	}

	if cmd.Get("KEY") != "" {
		t.Errorf("KEY = %v, want empty string", cmd.Get("KEY"))
	}
}

func TestParser_ParseCommand_EOF(t *testing.T) {
	input := ""
	parser := NewParser(strings.NewReader(input))

	_, err := parser.ParseCommand()
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestParser_ParseCommand_MultipleCommands(t *testing.T) {
	input := "HELLO VERSION\nPING\nQUIT\n"
	parser := NewParser(strings.NewReader(input))

	// First command
	cmd1, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("First ParseCommand() error = %v", err)
	}
	if cmd1.Verb != "HELLO" {
		t.Errorf("First command Verb = %v, want HELLO", cmd1.Verb)
	}

	// Second command
	cmd2, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Second ParseCommand() error = %v", err)
	}
	if cmd2.Verb != "PING" {
		t.Errorf("Second command Verb = %v, want PING", cmd2.Verb)
	}

	// Third command
	cmd3, err := parser.ParseCommand()
	if err != nil {
		t.Fatalf("Third ParseCommand() error = %v", err)
	}
	if cmd3.Verb != "QUIT" {
		t.Errorf("Third command Verb = %v, want QUIT", cmd3.Verb)
	}
}

func TestParser_ParseCommand_UnclosedQuote(t *testing.T) {
	input := "TEST KEY=\"unclosed\n"
	parser := NewParser(strings.NewReader(input))

	_, err := parser.ParseCommand()
	if err == nil {
		t.Error("Expected error for unclosed quote")
	}
}

func TestParser_ParseCommand_TrailingBackslash(t *testing.T) {
	input := "TEST KEY=value\\\n"
	parser := NewParser(strings.NewReader(input))

	_, err := parser.ParseCommand()
	if err == nil {
		t.Error("Expected error for trailing backslash")
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "simple options",
			input: "KEY1=value1 KEY2=value2",
			want:  map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:  "quoted value",
			input: "NAME=\"test value\"",
			want:  map[string]string{"NAME": "test value"},
		},
		{
			name:  "empty value",
			input: "KEY=",
			want:  map[string]string{"KEY": ""},
		},
		{
			name:    "invalid format",
			input:   "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOptions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("ParseOptions()[%v] = %v, want %v", k, got[k], v)
					}
				}
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "simple tokens",
			input: "HELLO VERSION MIN=3.0",
			want:  []string{"HELLO", "VERSION", "MIN=3.0"},
		},
		{
			name:  "quoted string",
			input: "KEY=\"value with spaces\"",
			want:  []string{"KEY=value with spaces"},
		},
		{
			name:  "escaped backslash",
			input: "KEY=value\\\\ test",
			want:  []string{"KEY=value\\", "test"},
		},
		{
			name:    "unclosed quote",
			input:   "KEY=\"unclosed",
			wantErr: true,
		},
		{
			name:    "trailing backslash",
			input:   "KEY=value\\",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("tokenize() len = %v, want %v", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("tokenize()[%v] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
