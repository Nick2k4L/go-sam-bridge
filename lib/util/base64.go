// Package util provides common utility functions for the SAM bridge.
package util

import (
	"fmt"

	"github.com/go-i2p/common/base64"
)

// EncodeDestinationBase64 encodes binary destination data to I2P Base64 format.
// This uses the I2P-specific Base64 alphabet (- instead of +, ~ instead of /)
// per the SAM and I2P specifications.
//
// The encoded string can be used in SAM protocol responses (DEST REPLY)
// and PrivateKeyFile storage.
//
// Example:
//
//	encoded := util.EncodeDestinationBase64(destinationBytes)
//	// Returns: "SGVsbG8-..." (I2P Base64 format)
func EncodeDestinationBase64(data []byte) string {
	return base64.EncodeToString(data)
}

// DecodeDestinationBase64 decodes an I2P Base64 string back to binary data.
// This handles the I2P-specific Base64 alphabet and validates the input.
//
// The input string should be in I2P Base64 format (using - and ~ instead
// of + and / respectively). Returns an error if the string contains
// invalid characters or malformed padding.
//
// Example:
//
//	data, err := util.DecodeDestinationBase64("SGVsbG8-...")
//	if err != nil {
//	    return err
//	}
func DecodeDestinationBase64(str string) ([]byte, error) {
	data, err := base64.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDestination, err)
	}
	return data, nil
}
