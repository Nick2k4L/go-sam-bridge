package protocol

import (
	"strings"
	"testing"
)

func TestResponse_Basic(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Response
		expected string
	}{
		{
			name: "simple verb only",
			build: func() *Response {
				return NewResponse("PONG")
			},
			expected: "PONG\n",
		},
		{
			name: "verb with action",
			build: func() *Response {
				return NewResponse("HELLO").WithAction("REPLY")
			},
			expected: "HELLO REPLY\n",
		},
		{
			name: "hello reply ok",
			build: func() *Response {
				return NewResponse("HELLO").
					WithAction("REPLY").
					WithResult("OK").
					WithVersion("3.3")
			},
			expected: "HELLO REPLY RESULT=OK VERSION=3.3\n",
		},
		{
			name: "session status with destination",
			build: func() *Response {
				return NewResponse("SESSION").
					WithAction("STATUS").
					WithResult("OK").
					WithDestination("abc123")
			},
			expected: "SESSION STATUS RESULT=OK DESTINATION=abc123\n",
		},
		{
			name: "error with message",
			build: func() *Response {
				return NewResponse("SESSION").
					WithAction("STATUS").
					WithResult("I2P_ERROR").
					WithMessage("connection failed")
			},
			expected: "SESSION STATUS RESULT=I2P_ERROR MESSAGE=\"connection failed\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.build().String()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResponse_Quoting(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "simple value no quotes",
			key:      "ID",
			value:    "test123",
			expected: "ID=test123",
		},
		{
			name:     "value with space",
			key:      "MESSAGE",
			value:    "hello world",
			expected: `MESSAGE="hello world"`,
		},
		{
			name:     "value with tab",
			key:      "MESSAGE",
			value:    "hello\tworld",
			expected: `MESSAGE="hello	world"`,
		},
		{
			name:     "value with quote",
			key:      "MESSAGE",
			value:    `say "hello"`,
			expected: `MESSAGE="say \"hello\""`,
		},
		{
			name:     "value with backslash",
			key:      "PATH",
			value:    `C:\test\path`,
			expected: `PATH="C:\\test\\path"`,
		},
		{
			name:     "value with quote and backslash",
			key:      "MESSAGE",
			value:    `he said "hello\" there`,
			expected: `MESSAGE="he said \"hello\\\" there"`,
		},
		{
			name:     "empty value",
			key:      "ID",
			value:    "",
			expected: "ID=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResponse("TEST").WithOption(tt.key, tt.value)
			result := r.String()

			if !strings.Contains(result, tt.expected) {
				t.Errorf("response %q should contain %q", result, tt.expected)
			}
		})
	}
}

func TestResponse_Bytes(t *testing.T) {
	r := NewResponse("HELLO").WithAction("REPLY").WithResult("OK")

	bytes := r.Bytes()
	str := r.String()

	if string(bytes) != str {
		t.Errorf("Bytes() = %q, String() = %q, should match", string(bytes), str)
	}
}

func TestHelperResponses(t *testing.T) {
	t.Run("HelloReplyOK", func(t *testing.T) {
		r := HelloReplyOK("3.3")
		expected := "HELLO REPLY RESULT=OK VERSION=3.3\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})

	t.Run("HelloReplyNoVersion", func(t *testing.T) {
		r := HelloReplyNoVersion()
		expected := "HELLO REPLY RESULT=NOVERSION\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})

	t.Run("HelloReplyError", func(t *testing.T) {
		r := HelloReplyError("test error")
		result := r.String()
		if !strings.Contains(result, "RESULT=I2P_ERROR") {
			t.Error("should contain I2P_ERROR result")
		}
		if !strings.Contains(result, "MESSAGE=") {
			t.Error("should contain MESSAGE")
		}
	})

	t.Run("SessionStatusOK", func(t *testing.T) {
		r := SessionStatusOK("testdest")
		expected := "SESSION STATUS RESULT=OK DESTINATION=testdest\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})

	t.Run("SessionStatusError", func(t *testing.T) {
		r := SessionStatusError("DUPLICATED_ID", "id already exists")
		result := r.String()
		if !strings.Contains(result, "RESULT=DUPLICATED_ID") {
			t.Error("should contain DUPLICATED_ID result")
		}
	})

	t.Run("StreamStatusOK", func(t *testing.T) {
		r := StreamStatusOK()
		expected := "STREAM STATUS RESULT=OK\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})

	t.Run("DestReply", func(t *testing.T) {
		r := DestReply("pubkey", "privkey")
		result := r.String()
		if !strings.Contains(result, "PUB=pubkey") {
			t.Error("should contain PUB=pubkey")
		}
		if !strings.Contains(result, "PRIV=privkey") {
			t.Error("should contain PRIV=privkey")
		}
	})

	t.Run("NamingReplyOK", func(t *testing.T) {
		r := NamingReplyOK("test.i2p", "destbase64")
		result := r.String()
		if !strings.Contains(result, "RESULT=OK") {
			t.Error("should contain RESULT=OK")
		}
		if !strings.Contains(result, "NAME=test.i2p") {
			t.Error("should contain NAME=test.i2p")
		}
	})

	t.Run("NamingReplyNotFound", func(t *testing.T) {
		r := NamingReplyNotFound("unknown.i2p")
		result := r.String()
		if !strings.Contains(result, "RESULT=KEY_NOT_FOUND") {
			t.Error("should contain KEY_NOT_FOUND result")
		}
	})

	t.Run("Pong with data", func(t *testing.T) {
		r := Pong("test data")
		expected := "PONG test data\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})

	t.Run("Pong empty", func(t *testing.T) {
		r := Pong("")
		expected := "PONG\n"
		if r.String() != expected {
			t.Errorf("got %q, want %q", r.String(), expected)
		}
	})
}
