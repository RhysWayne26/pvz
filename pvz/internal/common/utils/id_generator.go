package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const signed63Mask = (1 << 63) - 1

func GenerateID() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, fmt.Errorf("failed to generate id: %w", err)
	}
	return binary.BigEndian.Uint64(b[:]) & signed63Mask, nil
}
