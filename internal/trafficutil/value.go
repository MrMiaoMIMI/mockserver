package trafficutil

import (
	"crypto/sha256"
	"encoding/binary"
)

func FieldValueHash(value string) uint64 {
	sum := sha256.Sum256([]byte(value))
	return binary.BigEndian.Uint64(sum[:8])
}

func FieldValuePreview(value string) string {
	runes := []rune(value)
	if len(runes) <= 255 {
		return value
	}
	return string(runes[:255])
}
