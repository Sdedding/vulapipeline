package validate

import (
	"encoding/base64"
	"fmt"
)

func (v *V) Base64Bytes(key string, value string, length int) []byte {
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		v.Errors = append(v.Errors, KeyError{
			Key:     key,
			Value:   value,
			Message: "invalid base64",
			Err:     err,
		})
		return b
	}

	if length >= 0 && len(b) != length {
		v.Errors = append(v.Errors, KeyError{
			Key:     key,
			Value:   value,
			Message: "invalid length",
			Err:     fmt.Errorf("expected %d bytes but got %d", length, len(b)),
		})
	}
	return b
}
