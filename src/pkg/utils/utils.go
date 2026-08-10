package utils

import (
	"apigw/src/pkg/uuid"
	"encoding/base64"
	"errors"
	"strconv"
	"unicode"
)

// CertToBase64 encodes raw PEM certificate text into Base64 string
func CertToBase64(pemRaw string) string {
	return base64.StdEncoding.EncodeToString([]byte(pemRaw))
}

// Base64ToCert decodes Base64 string and restore original PEM certificate content
func Base64ToCert(b64Str string) ([]byte, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return nil, errors.New("base64 decode failed, invalid certificate encoding")
	}
	return rawBytes, nil
}

// Str2Int converts a string to an integer.
func Str2Int(str string) (int, error) {
	n, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// CreateUuid generates a new UUID v4
// Returns a string representation of a version 4 UUID
func CreateUuid() string {
	return uuid.V4UUID()
}

func CheckUuid(u string) bool {
	return IsValid32UUID(u)
}

// IsValid32UUID checks if a string is a valid 32-character UUID (without hyphens)
// Returns true if the string is exactly 32 characters long
// and contains only valid hexadecimal characters (0-9, a-f, A-F)
func IsValid32UUID(uuid string) bool {
	// First check if length is exactly 32 characters
	if len(uuid) != 32 {
		return false
	}

	// Check each character is a valid hexadecimal digit
	for _, c := range uuid {
		if !unicode.Is(unicode.Hex_Digit, c) {
			return false
		}
	}

	return true
}
