package util

import (
	"bytes"
	"testing"
)

func TestEncodeDestinationBase64(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty input",
			input: []byte{},
			want:  "",
		},
		{
			name:  "simple text",
			input: []byte("Hello"),
			want:  "SGVsbG8=",
		},
		{
			name:  "binary data",
			input: []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			want:  "AAECA~8=", // I2P uses ~ not /
		},
		{
			name:  "I2P uses ~ not /",
			input: []byte{0xFF, 0xFF, 0xFF},
			want:  "~~~~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeDestinationBase64(tt.input)
			if got != tt.want {
				t.Errorf("EncodeDestinationBase64() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeDestinationBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "valid encoding",
			input:   "SGVsbG8=",
			want:    []byte("Hello"),
			wantErr: false,
		},
		{
			name:    "binary data roundtrip",
			input:   "AAECA~8=",
			want:    []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			wantErr: false,
		},
		{
			name:    "invalid characters",
			input:   "!!!invalid!!!",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "tilde character (I2P-specific)",
			input:   "~~~~",
			want:    []byte{0xFF, 0xFF, 0xFF},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeDestinationBase64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeDestinationBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.want) {
				t.Errorf("DecodeDestinationBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase64Roundtrip(t *testing.T) {
	testData := [][]byte{
		{},
		{0x00},
		[]byte("Hello World"),
		{0xFF, 0xFE, 0xFD, 0xFC},
		make([]byte, 256), // Larger data
	}

	for i, data := range testData {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Encode
			encoded := EncodeDestinationBase64(data)

			// Decode
			decoded, err := DecodeDestinationBase64(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			// Verify roundtrip
			if !bytes.Equal(data, decoded) {
				t.Errorf("Roundtrip failed: got %v, want %v", decoded, data)
			}
		})
	}
}
