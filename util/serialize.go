package util

import "encoding/binary"

func SerializeInt64(val int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(val))
	return buf
}

func DeserializeInt64(val []byte) int64 {
	return int64(binary.LittleEndian.Uint64(val))
}
