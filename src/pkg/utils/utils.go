package utils

import (
	"encoding/base64"
	"errors"
)

// CertToBase64 encodes raw PEM certificate text into Base64 string
func CertToBase64(pemRaw string) string {
	return base64.StdEncoding.EncodeToString([]byte(pemRaw))
}

// Base64ToCert decodes Base64 string and restore original PEM certificate content
func Base64ToCert(b64Str string) (string, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", errors.New("base64 decode failed, invalid certificate encoding")
	}
	return string(rawBytes), nil
}
