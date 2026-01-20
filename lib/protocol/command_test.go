package protocol

import (
	"testing"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand("SESSION", "CREATE")

	if cmd.Verb != "SESSION" {
		t.Errorf("Verb = %q, want %q", cmd.Verb, "SESSION")
	}
	if cmd.Action != "CREATE" {
		t.Errorf("Action = %q, want %q", cmd.Action, "CREATE")
	}
	if cmd.Options == nil {
		t.Error("Options should not be nil")
	}
}

func TestCommand_Get(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		key      string
		expected string
	}{
		{
			name:     "existing key",
			options:  map[string]string{"ID": "test123"},
			key:      "ID",
			expected: "test123",
		},
		{
			name:     "missing key",
			options:  map[string]string{"ID": "test123"},
			key:      "STYLE",
			expected: "",
		},
		{
			name:     "empty value",
			options:  map[string]string{"SILENT": ""},
			key:      "SILENT",
			expected: "",
		},
		{
			name:     "nil options",
			options:  nil,
			key:      "ID",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Options: tt.options}
			got := cmd.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestCommand_GetOrDefault(t *testing.T) {
	tests := []struct {
		name       string
		options    map[string]string
		key        string
		defaultVal string
		expected   string
	}{
		{
			name:       "existing key",
			options:    map[string]string{"ID": "test123"},
			key:        "ID",
			defaultVal: "default",
			expected:   "test123",
		},
		{
			name:       "missing key uses default",
			options:    map[string]string{"ID": "test123"},
			key:        "STYLE",
			defaultVal: "STREAM",
			expected:   "STREAM",
		},
		{
			name:       "empty value returns empty not default",
			options:    map[string]string{"SILENT": ""},
			key:        "SILENT",
			defaultVal: "true",
			expected:   "",
		},
		{
			name:       "nil options uses default",
			options:    nil,
			key:        "ID",
			defaultVal: "default",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Options: tt.options}
			got := cmd.GetOrDefault(tt.key, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("GetOrDefault(%q, %q) = %q, want %q",
					tt.key, tt.defaultVal, got, tt.expected)
			}
		})
	}
}

func TestCommand_Has(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		key      string
		expected bool
	}{
		{
			name:     "existing key with value",
			options:  map[string]string{"ID": "test123"},
			key:      "ID",
			expected: true,
		},
		{
			name:     "existing key with empty value",
			options:  map[string]string{"SILENT": ""},
			key:      "SILENT",
			expected: true,
		},
		{
			name:     "missing key",
			options:  map[string]string{"ID": "test123"},
			key:      "STYLE",
			expected: false,
		},
		{
			name:     "nil options",
			options:  nil,
			key:      "ID",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{Options: tt.options}
			got := cmd.Has(tt.key)
			if got != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestCommand_Set(t *testing.T) {
	t.Run("set on initialized options", func(t *testing.T) {
		cmd := NewCommand("SESSION", "CREATE")
		cmd.Set("ID", "test123")

		if cmd.Get("ID") != "test123" {
			t.Error("Set failed to add option")
		}
	})

	t.Run("set on nil options", func(t *testing.T) {
		cmd := &Command{}
		cmd.Set("ID", "test123")

		if cmd.Options == nil {
			t.Error("Set should initialize Options")
		}
		if cmd.Get("ID") != "test123" {
			t.Error("Set failed to add option")
		}
	})

	t.Run("overwrite existing value", func(t *testing.T) {
		cmd := NewCommand("SESSION", "CREATE")
		cmd.Set("ID", "old")
		cmd.Set("ID", "new")

		if cmd.Get("ID") != "new" {
			t.Error("Set failed to overwrite option")
		}
	})
}
