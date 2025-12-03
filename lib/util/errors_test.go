package util

import (
	"errors"
	"testing"
)

func TestSessionError(t *testing.T) {
	baseErr := errors.New("test")
	err := NewSessionError("s1", "op", baseErr)
	if err.SessionID != "s1" {
		t.Error("wrong ID")
	}
}
