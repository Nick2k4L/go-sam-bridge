package protocol

import (
	"testing"
)

func TestCommand_Get(t *testing.T) {
	cmd := &Command{
		Options: map[string]string{"KEY": "value"},
	}

	if got := cmd.Get("KEY"); got != "value" {
		t.Errorf("Get(KEY) = %v, want value", got)
	}

	if got := cmd.Get("MISSING"); got != "" {
		t.Errorf("Get(MISSING) = %v, want empty", got)
	}
}

func TestCommand_GetOr(t *testing.T) {
	cmd := &Command{
		Options: map[string]string{"KEY": "value"},
	}

	if got := cmd.GetOr("KEY", "default"); got != "value" {
		t.Errorf("GetOr(KEY) = %v, want value", got)
	}

	if got := cmd.GetOr("MISSING", "default"); got != "default" {
		t.Errorf("GetOr(MISSING) = %v, want default", got)
	}
}

func TestCommand_Has(t *testing.T) {
	cmd := &Command{
		Options: map[string]string{"KEY": ""},
	}

	if !cmd.Has("KEY") {
		t.Error("Has(KEY) should be true")
	}

	if cmd.Has("MISSING") {
		t.Error("Has(MISSING) should be false")
	}
}

func TestCommand_FullCommand(t *testing.T) {
	tests := []struct {
		verb   string
		action string
		want   string
	}{
		{"SESSION", "CREATE", "SESSION CREATE"},
		{"PING", "", "PING"},
		{"DEST", "GENERATE", "DEST GENERATE"},
	}

	for _, tt := range tests {
		cmd := &Command{Verb: tt.verb, Action: tt.action}
		if got := cmd.FullCommand(); got != tt.want {
			t.Errorf("FullCommand() = %v, want %v", got, tt.want)
		}
	}
}
