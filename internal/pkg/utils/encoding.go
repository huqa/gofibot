package utils

import "encoding/binary"

// Itob converts an int to an 8-byte big endian encoded byte slice
func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Btoi converts an 8-byte big endian encoded byte slice to an int
func Btoi(v []byte) int {
	return int(binary.BigEndian.Uint64(v))
}
