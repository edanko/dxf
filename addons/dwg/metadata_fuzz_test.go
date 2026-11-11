package dwg

import (
	"encoding/binary"
	"testing"
)

func FuzzReadMetadata(f *testing.F) {
	seed := make([]byte, 0x19)
	copy(seed, "AC1015")
	binary.LittleEndian.PutUint16(seed[0x13:0x15], 30)
	binary.LittleEndian.PutUint32(seed[0x15:0x19], 0)
	f.Add(seed)
	f.Add([]byte("AC1021"))
	f.Add([]byte("not a dwg"))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ReadMetadata(data)
	})
}
