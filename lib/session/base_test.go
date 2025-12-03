package session

import "testing"

func TestSessionStyle_String(t *testing.T) {
	if StyleStream.String() != "STREAM" {
		t.Errorf("StyleStream.String() = %v, want STREAM", StyleStream.String())
	}

	if StyleDatagram.String() != "DATAGRAM" {
		t.Errorf("StyleDatagram.String() = %v, want DATAGRAM", StyleDatagram.String())
	}
}

func TestSessionStatus_String(t *testing.T) {
	if StatusReady.String() != "READY" {
		t.Errorf("StatusReady.String() = %v, want READY", StatusReady.String())
	}
}

func TestNewBaseSession(t *testing.T) {
	config := &SessionConfig{
		ID:    "test1",
		Style: StyleStream,
	}

	base := NewBaseSession(config, nil, nil)

	if base.ID() != "test1" {
		t.Errorf("ID() = %v, want test1", base.ID())
	}

	if base.Style() != StyleStream {
		t.Errorf("Style() = %v, want StyleStream", base.Style())
	}
}

func TestBaseSession_Close(t *testing.T) {
	config := &SessionConfig{ID: "close-test", Style: StyleStream}
	base := NewBaseSession(config, nil, nil)

	if base.IsClosed() {
		t.Error("New session should not be closed")
	}

	err := base.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !base.IsClosed() {
		t.Error("Session should be closed after Close()")
	}
}
