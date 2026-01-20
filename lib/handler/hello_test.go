package handler

import (
	"strings"
	"testing"

	"github.com/go-i2p/go-sam-bridge/lib/protocol"
)

func TestHelloHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		config         HelloConfig
		command        *protocol.Command
		handshakeDone  bool
		wantResult     string
		wantVersion    string
		wantHandshake  bool
		checkNoVersion bool
	}{
		{
			name:   "basic hello with defaults",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:    "HELLO",
				Action:  "VERSION",
				Options: map[string]string{},
			},
			wantResult:    protocol.ResultOK,
			wantVersion:   protocol.SAMVersionMax,
			wantHandshake: true,
		},
		{
			name:   "hello with MIN and MAX",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "3.1",
					"MAX": "3.1",
				},
			},
			wantResult:    protocol.ResultOK,
			wantVersion:   "3.1",
			wantHandshake: true,
		},
		{
			name:   "hello with MIN only",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "3.2",
				},
			},
			wantResult:    protocol.ResultOK,
			wantVersion:   "3.3",
			wantHandshake: true,
		},
		{
			name:   "hello with MAX only",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MAX": "3.2",
				},
			},
			wantResult:    protocol.ResultOK,
			wantVersion:   "3.2",
			wantHandshake: true,
		},
		{
			name:   "no compatible version - client too old",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "2.0",
					"MAX": "2.9",
				},
			},
			wantResult:     protocol.ResultNoVersion,
			checkNoVersion: true,
		},
		{
			name:   "no compatible version - client too new",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "4.0",
					"MAX": "4.1",
				},
			},
			wantResult:     protocol.ResultNoVersion,
			checkNoVersion: true,
		},
		{
			name:   "invalid MIN version format",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "invalid",
				},
			},
			wantResult: protocol.ResultI2PError,
		},
		{
			name:   "invalid MAX version format",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MAX": "3.x",
				},
			},
			wantResult: protocol.ResultI2PError,
		},
		{
			name:   "MIN greater than MAX",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "3.3",
					"MAX": "3.1",
				},
			},
			wantResult: protocol.ResultI2PError,
		},
		{
			name:          "hello already completed",
			config:        DefaultHelloConfig(),
			handshakeDone: true,
			command: &protocol.Command{
				Verb:    "HELLO",
				Action:  "VERSION",
				Options: map[string]string{},
			},
			wantResult: protocol.ResultI2PError,
		},
		{
			name: "restricted server version range",
			config: HelloConfig{
				MinVersion:  "3.1",
				MaxVersion:  "3.2",
				RequireAuth: false,
			},
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"MIN": "3.0",
					"MAX": "3.3",
				},
			},
			wantResult:    protocol.ResultOK,
			wantVersion:   "3.2",
			wantHandshake: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHelloHandler(tt.config)
			ctx := NewContext(&mockConn{}, nil)
			ctx.HandshakeComplete = tt.handshakeDone

			resp, err := handler.Handle(ctx, tt.command)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Handle() returned nil response")
			}

			respStr := resp.String()
			if !containsResult(respStr, tt.wantResult) {
				t.Errorf("Handle() result = %q, want RESULT=%s", respStr, tt.wantResult)
			}

			if tt.wantVersion != "" {
				if !containsVersion(respStr, tt.wantVersion) {
					t.Errorf("Handle() = %q, want VERSION=%s", respStr, tt.wantVersion)
				}
			}

			if tt.wantHandshake && !ctx.HandshakeComplete {
				t.Error("Handle() did not set HandshakeComplete")
			}

			if tt.wantHandshake && ctx.Version != tt.wantVersion {
				t.Errorf("Handle() ctx.Version = %q, want %q", ctx.Version, tt.wantVersion)
			}
		})
	}
}

func TestHelloHandler_Authentication(t *testing.T) {
	validAuth := func(user, password string) bool {
		return user == "testuser" && password == "testpass"
	}

	tests := []struct {
		name       string
		config     HelloConfig
		command    *protocol.Command
		wantResult string
		wantAuth   bool
	}{
		{
			name: "auth required and provided correctly",
			config: HelloConfig{
				MinVersion:  "3.0",
				MaxVersion:  "3.3",
				RequireAuth: true,
				AuthFunc:    validAuth,
			},
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"USER":     "testuser",
					"PASSWORD": "testpass",
				},
			},
			wantResult: protocol.ResultOK,
			wantAuth:   true,
		},
		{
			name: "auth required but wrong password",
			config: HelloConfig{
				MinVersion:  "3.0",
				MaxVersion:  "3.3",
				RequireAuth: true,
				AuthFunc:    validAuth,
			},
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"USER":     "testuser",
					"PASSWORD": "wrongpass",
				},
			},
			wantResult: protocol.ResultI2PError,
			wantAuth:   false,
		},
		{
			name: "auth required but missing credentials",
			config: HelloConfig{
				MinVersion:  "3.0",
				MaxVersion:  "3.3",
				RequireAuth: true,
				AuthFunc:    validAuth,
			},
			command: &protocol.Command{
				Verb:    "HELLO",
				Action:  "VERSION",
				Options: map[string]string{},
			},
			wantResult: protocol.ResultI2PError,
			wantAuth:   false,
		},
		{
			name: "auth required missing password",
			config: HelloConfig{
				MinVersion:  "3.0",
				MaxVersion:  "3.3",
				RequireAuth: true,
				AuthFunc:    validAuth,
			},
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"USER": "testuser",
				},
			},
			wantResult: protocol.ResultI2PError,
			wantAuth:   false,
		},
		{
			name:   "auth not required with credentials",
			config: DefaultHelloConfig(),
			command: &protocol.Command{
				Verb:   "HELLO",
				Action: "VERSION",
				Options: map[string]string{
					"USER":     "testuser",
					"PASSWORD": "testpass",
				},
			},
			wantResult: protocol.ResultOK,
			wantAuth:   false, // Auth not required, so not marked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHelloHandler(tt.config)
			ctx := NewContext(&mockConn{}, nil)

			resp, err := handler.Handle(ctx, tt.command)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			respStr := resp.String()
			if !containsResult(respStr, tt.wantResult) {
				t.Errorf("Handle() = %q, want RESULT=%s", respStr, tt.wantResult)
			}

			if ctx.Authenticated != tt.wantAuth {
				t.Errorf("Handle() ctx.Authenticated = %v, want %v", ctx.Authenticated, tt.wantAuth)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"3.0", "3.0", 0},
		{"3.0", "3.1", -1},
		{"3.1", "3.0", 1},
		{"3.0", "4.0", -1},
		{"4.0", "3.0", 1},
		{"3.10", "3.2", 1},  // 10 > 2
		{"3.2", "3.10", -1}, // 2 < 10
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"3.0", true},
		{"3.1", true},
		{"3.3", true},
		{"10.20", true},
		{"3", false},
		{"3.0.1", false},
		{"3.x", false},
		{"x.0", false},
		{"", false},
		{".", false},
		{"a.b", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isValidVersion(tt.version)
			if got != tt.want {
				t.Errorf("isValidVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestLaterEarlierVersion(t *testing.T) {
	tests := []struct {
		a, b        string
		wantLater   string
		wantEarlier string
	}{
		{"3.0", "3.1", "3.1", "3.0"},
		{"3.3", "3.1", "3.3", "3.1"},
		{"3.2", "3.2", "3.2", "3.2"},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			gotLater := laterVersion(tt.a, tt.b)
			if gotLater != tt.wantLater {
				t.Errorf("laterVersion(%q, %q) = %q, want %q", tt.a, tt.b, gotLater, tt.wantLater)
			}

			gotEarlier := earlierVersion(tt.a, tt.b)
			if gotEarlier != tt.wantEarlier {
				t.Errorf("earlierVersion(%q, %q) = %q, want %q", tt.a, tt.b, gotEarlier, tt.wantEarlier)
			}
		})
	}
}

func TestHelloResponses(t *testing.T) {
	t.Run("helloOK", func(t *testing.T) {
		resp := helloOK("3.2")
		got := resp.String()
		want := "HELLO REPLY RESULT=OK VERSION=3.2\n"
		if got != want {
			t.Errorf("helloOK() = %q, want %q", got, want)
		}
	})

	t.Run("helloNoVersion", func(t *testing.T) {
		resp := helloNoVersion()
		got := resp.String()
		want := "HELLO REPLY RESULT=NOVERSION\n"
		if got != want {
			t.Errorf("helloNoVersion() = %q, want %q", got, want)
		}
	})

	t.Run("helloError", func(t *testing.T) {
		resp := helloError("test error")
		got := resp.String()
		if !containsResult(got, protocol.ResultI2PError) {
			t.Errorf("helloError() = %q, want RESULT=I2P_ERROR", got)
		}
		if !contains(got, "MESSAGE=") {
			t.Errorf("helloError() = %q, want MESSAGE=", got)
		}
	})
}

func TestDefaultHelloConfig(t *testing.T) {
	cfg := DefaultHelloConfig()

	if cfg.MinVersion != protocol.SAMVersionMin {
		t.Errorf("MinVersion = %q, want %q", cfg.MinVersion, protocol.SAMVersionMin)
	}
	if cfg.MaxVersion != protocol.SAMVersionMax {
		t.Errorf("MaxVersion = %q, want %q", cfg.MaxVersion, protocol.SAMVersionMax)
	}
	if cfg.RequireAuth {
		t.Error("RequireAuth should be false by default")
	}
	if cfg.AuthFunc != nil {
		t.Error("AuthFunc should be nil by default")
	}
}

func TestVersionError(t *testing.T) {
	err := &versionError{msg: "test error message"}
	if err.Error() != "test error message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error message")
	}
}

// Helper functions for test assertions
func containsResult(s, result string) bool {
	return strings.Contains(s, "RESULT="+result)
}

func containsVersion(s, version string) bool {
	return strings.Contains(s, "VERSION="+version)
}
