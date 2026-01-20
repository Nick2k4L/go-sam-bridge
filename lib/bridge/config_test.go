package bridge

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ListenAddr != DefaultListenAddr {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, DefaultListenAddr)
	}
	if cfg.I2CPAddr != DefaultI2CPAddr {
		t.Errorf("I2CPAddr = %q, want %q", cfg.I2CPAddr, DefaultI2CPAddr)
	}
	if cfg.DatagramPort != DefaultDatagramPort {
		t.Errorf("DatagramPort = %d, want %d", cfg.DatagramPort, DefaultDatagramPort)
	}
	if cfg.TLSConfig != nil {
		t.Errorf("TLSConfig = %v, want nil", cfg.TLSConfig)
	}
	if cfg.Auth.Required {
		t.Error("Auth.Required = true, want false")
	}
	if cfg.Auth.Users == nil {
		t.Error("Auth.Users = nil, want empty map")
	}
	if cfg.Timeouts.Handshake != DefaultHandshakeTimeout {
		t.Errorf("Timeouts.Handshake = %v, want %v", cfg.Timeouts.Handshake, DefaultHandshakeTimeout)
	}
	if cfg.Timeouts.Command != DefaultCommandTimeout {
		t.Errorf("Timeouts.Command = %v, want %v", cfg.Timeouts.Command, DefaultCommandTimeout)
	}
	if cfg.Limits.ReadBufferSize != DefaultReadBufferSize {
		t.Errorf("Limits.ReadBufferSize = %d, want %d", cfg.Limits.ReadBufferSize, DefaultReadBufferSize)
	}
	if cfg.Limits.MaxLineLength != DefaultMaxLineLength {
		t.Errorf("Limits.MaxLineLength = %d, want %d", cfg.Limits.MaxLineLength, DefaultMaxLineLength)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*Config)
		wantErr   bool
		wantField string
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:      "empty listen address",
			modify:    func(c *Config) { c.ListenAddr = "" },
			wantErr:   true,
			wantField: "ListenAddr",
		},
		{
			name:      "empty I2CP address",
			modify:    func(c *Config) { c.I2CPAddr = "" },
			wantErr:   true,
			wantField: "I2CPAddr",
		},
		{
			name:      "negative datagram port",
			modify:    func(c *Config) { c.DatagramPort = -1 },
			wantErr:   true,
			wantField: "DatagramPort",
		},
		{
			name:      "datagram port too high",
			modify:    func(c *Config) { c.DatagramPort = 65536 },
			wantErr:   true,
			wantField: "DatagramPort",
		},
		{
			name:    "datagram port zero (disabled)",
			modify:  func(c *Config) { c.DatagramPort = 0 },
			wantErr: false,
		},
		{
			name:    "datagram port max",
			modify:  func(c *Config) { c.DatagramPort = 65535 },
			wantErr: false,
		},
		{
			name:      "negative handshake timeout",
			modify:    func(c *Config) { c.Timeouts.Handshake = -1 * time.Second },
			wantErr:   true,
			wantField: "Timeouts.Handshake",
		},
		{
			name:    "zero handshake timeout",
			modify:  func(c *Config) { c.Timeouts.Handshake = 0 },
			wantErr: false,
		},
		{
			name:      "negative command timeout",
			modify:    func(c *Config) { c.Timeouts.Command = -1 * time.Second },
			wantErr:   true,
			wantField: "Timeouts.Command",
		},
		{
			name:      "zero read buffer size",
			modify:    func(c *Config) { c.Limits.ReadBufferSize = 0 },
			wantErr:   true,
			wantField: "Limits.ReadBufferSize",
		},
		{
			name:      "negative read buffer size",
			modify:    func(c *Config) { c.Limits.ReadBufferSize = -1 },
			wantErr:   true,
			wantField: "Limits.ReadBufferSize",
		},
		{
			name:      "zero max line length",
			modify:    func(c *Config) { c.Limits.MaxLineLength = 0 },
			wantErr:   true,
			wantField: "Limits.MaxLineLength",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() = nil, want error")
					return
				}
				cfgErr, ok := err.(*ConfigError)
				if !ok {
					t.Errorf("error type = %T, want *ConfigError", err)
					return
				}
				if cfgErr.Field != tt.wantField {
					t.Errorf("error field = %q, want %q", cfgErr.Field, tt.wantField)
				}
			} else if err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestConfig_WithListenAddr(t *testing.T) {
	cfg := DefaultConfig()
	newCfg := cfg.WithListenAddr("127.0.0.1:8080")

	if cfg.ListenAddr == "127.0.0.1:8080" {
		t.Error("original config was modified")
	}
	if newCfg.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("ListenAddr = %q, want %q", newCfg.ListenAddr, "127.0.0.1:8080")
	}
}

func TestConfig_WithI2CPAddr(t *testing.T) {
	cfg := DefaultConfig()
	newCfg := cfg.WithI2CPAddr("192.168.1.1:7654")

	if cfg.I2CPAddr == "192.168.1.1:7654" {
		t.Error("original config was modified")
	}
	if newCfg.I2CPAddr != "192.168.1.1:7654" {
		t.Errorf("I2CPAddr = %q, want %q", newCfg.I2CPAddr, "192.168.1.1:7654")
	}
}

func TestConfig_WithTLS(t *testing.T) {
	cfg := DefaultConfig()
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	newCfg := cfg.WithTLS(tlsConfig)

	if cfg.TLSConfig != nil {
		t.Error("original config was modified")
	}
	if newCfg.TLSConfig != tlsConfig {
		t.Error("TLSConfig was not set correctly")
	}
}

func TestConfig_WithAuth(t *testing.T) {
	cfg := DefaultConfig()
	users := map[string]string{"admin": "secret"}
	newCfg := cfg.WithAuth(true, users)

	if cfg.Auth.Required {
		t.Error("original config was modified")
	}
	if !newCfg.Auth.Required {
		t.Error("Auth.Required = false, want true")
	}
	if newCfg.Auth.Users["admin"] != "secret" {
		t.Error("Auth.Users was not set correctly")
	}
}

func TestConfig_AddUser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddUser("testuser", "testpass")

	if cfg.Auth.Users["testuser"] != "testpass" {
		t.Error("user was not added")
	}
}

func TestConfig_AddUser_NilMap(t *testing.T) {
	cfg := &Config{}
	cfg.AddUser("testuser", "testpass")

	if cfg.Auth.Users["testuser"] != "testpass" {
		t.Error("user was not added with nil map")
	}
}

func TestConfig_RemoveUser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddUser("testuser", "testpass")
	cfg.RemoveUser("testuser")

	if _, ok := cfg.Auth.Users["testuser"]; ok {
		t.Error("user was not removed")
	}
}

func TestConfig_RemoveUser_NotExists(t *testing.T) {
	cfg := DefaultConfig()
	// Should not panic
	cfg.RemoveUser("nonexistent")
}

func TestConfig_HasUser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddUser("testuser", "testpass")

	if !cfg.HasUser("testuser") {
		t.Error("HasUser(testuser) = false, want true")
	}
	if cfg.HasUser("nonexistent") {
		t.Error("HasUser(nonexistent) = true, want false")
	}
}

func TestConfig_CheckPassword(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddUser("testuser", "correctpass")

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			username: "testuser",
			password: "correctpass",
			want:     true,
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "wrongpass",
			want:     false,
		},
		{
			name:     "nonexistent user",
			username: "nobody",
			password: "anypass",
			want:     false,
		},
		{
			name:     "empty password when set",
			username: "testuser",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.CheckPassword(tt.username, tt.password)
			if got != tt.want {
				t.Errorf("CheckPassword(%q, %q) = %v, want %v",
					tt.username, tt.password, got, tt.want)
			}
		})
	}
}

func TestConfig_CheckPassword_EmptyPassword(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddUser("emptyuser", "")

	// Empty password should match empty password
	if !cfg.CheckPassword("emptyuser", "") {
		t.Error("CheckPassword with empty password should match")
	}
	if cfg.CheckPassword("emptyuser", "notempty") {
		t.Error("CheckPassword should not match non-empty password")
	}
}

func TestConfigError_Error(t *testing.T) {
	err := &ConfigError{
		Field:   "TestField",
		Message: "test message",
	}

	expected := "config error: TestField test message"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestConstants(t *testing.T) {
	// Verify default constants match SAM specification
	if DefaultListenAddr != ":7656" {
		t.Errorf("DefaultListenAddr = %q, want %q", DefaultListenAddr, ":7656")
	}
	if DefaultI2CPAddr != "127.0.0.1:7654" {
		t.Errorf("DefaultI2CPAddr = %q, want %q", DefaultI2CPAddr, "127.0.0.1:7654")
	}
	if DefaultDatagramPort != 7655 {
		t.Errorf("DefaultDatagramPort = %d, want %d", DefaultDatagramPort, 7655)
	}
}
