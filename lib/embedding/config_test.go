package embedding

import (
	"testing"
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

	if cfg.AuthUsers == nil {
		t.Error("AuthUsers should be initialized, not nil")
	}

	if cfg.Debug != false {
		t.Error("Debug should default to false")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr error
	}{
		{
			name:    "valid default config",
			cfg:     DefaultConfig(),
			wantErr: nil,
		},
		{
			name: "missing listen address",
			cfg: &Config{
				ListenAddr: "",
				I2CPAddr:   DefaultI2CPAddr,
			},
			wantErr: ErrMissingListenAddr,
		},
		{
			name: "missing I2CP address",
			cfg: &Config{
				ListenAddr: DefaultListenAddr,
				I2CPAddr:   "",
			},
			wantErr: ErrMissingI2CPAddr,
		},
		{
			name: "custom listener allows empty address",
			cfg: &Config{
				ListenAddr: "",
				I2CPAddr:   DefaultI2CPAddr,
				Listener:   &mockListener{},
			},
			wantErr: nil,
		},
		{
			name: "custom I2CP provider allows empty address",
			cfg: &Config{
				ListenAddr:   DefaultListenAddr,
				I2CPAddr:     "",
				I2CPProvider: &mockI2CPProvider{},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigToBridgeConfig(t *testing.T) {
	cfg := &Config{
		ListenAddr:   ":8000",
		I2CPAddr:     "10.0.0.1:7654",
		DatagramPort: 8001,
		AuthUsers: map[string]string{
			"user1": "pass1",
			"user2": "pass2",
		},
	}

	bridgeCfg := cfg.toBridgeConfig()

	if bridgeCfg.ListenAddr != cfg.ListenAddr {
		t.Errorf("ListenAddr = %q, want %q", bridgeCfg.ListenAddr, cfg.ListenAddr)
	}

	if bridgeCfg.I2CPAddr != cfg.I2CPAddr {
		t.Errorf("I2CPAddr = %q, want %q", bridgeCfg.I2CPAddr, cfg.I2CPAddr)
	}

	if bridgeCfg.DatagramPort != cfg.DatagramPort {
		t.Errorf("DatagramPort = %d, want %d", bridgeCfg.DatagramPort, cfg.DatagramPort)
	}

	if !bridgeCfg.Auth.Required {
		t.Error("Auth.Required should be true when users are configured")
	}

	if len(bridgeCfg.Auth.Users) != 2 {
		t.Errorf("Auth.Users length = %d, want 2", len(bridgeCfg.Auth.Users))
	}
}
