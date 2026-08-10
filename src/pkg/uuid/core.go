package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// UUID represents a 128-bit UUID (Universally Unique Identifier)
type UUID [16]byte

// String returns the UUID in standard 8-4-4-4-12 format
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(u[0:4]),
		binary.BigEndian.Uint16(u[4:6]),
		binary.BigEndian.Uint16(u[6:8]),
		binary.BigEndian.Uint16(u[8:10]),
		u[10:16])
}

// NoHyphenString returns the UUID without hyphens
func (u UUID) NoHyphenString() string {
	return fmt.Sprintf("%08x%04x%04x%04x%012x",
		binary.BigEndian.Uint32(u[0:4]),
		binary.BigEndian.Uint16(u[4:6]),
		binary.BigEndian.Uint16(u[6:8]),
		binary.BigEndian.Uint16(u[8:10]),
		u[10:16])
}

// Version returns the UUID version number (1-5)
func (u UUID) Version() uint8 {
	return (u[6] >> 4) & 0x0F
}

// Variant returns the UUID variant
func (u UUID) Variant() uint8 {
	return (u[8] >> 6) & 0x03
}

// IsValid checks if the UUID conforms to RFC 4122 standards
func (u UUID) IsValid() bool {
	version := u.Version()
	variant := u.Variant()

	// Valid versions are 1-5, valid RFC 4122 variant is 2
	return (version >= 1 && version <= 5) && variant == 2
}

// Predefined namespaces as specified in RFC 4122
var (
	NamespaceDNS  = UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceURL  = UUID{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceOID  = UUID{0x6b, 0xa7, 0xb8, 0x12, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceX500 = UUID{0x6b, 0xa7, 0xb8, 0x14, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
)

// NewV4 generates a version 4 UUID (random-based)
func NewV4() (UUID, error) {
	var u UUID
	if _, err := rand.Read(u[:]); err != nil {
		return UUID{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Set version to 4 (0100 in the most significant 4 bits of the 7th byte)
	u[6] = (u[6] & 0x0F) | 0x40

	// Set variant to RFC 4122 (10 in the most significant 2 bits of the 9th byte)
	u[8] = (u[8] & 0x3F) | 0x80

	return u, nil
}

// NewV4Batch generates multiple version 4 UUIDs in batch
func NewV4Batch(count int) ([]UUID, error) {
	if count <= 0 {
		return nil, fmt.Errorf("invalid count: %d", count)
	}

	batch := make([]UUID, count)
	data := make([]byte, 16*count)

	if _, err := rand.Read(data); err != nil {
		return nil, fmt.Errorf("failed to generate batch random bytes: %w", err)
	}

	for i := 0; i < count; i++ {
		copy(batch[i][:], data[i*16:(i+1)*16])

		// Set version and variant for each UUID
		batch[i][6] = (batch[i][6] & 0x0F) | 0x40
		batch[i][8] = (batch[i][8] & 0x3F) | 0x80
	}

	return batch, nil
}
