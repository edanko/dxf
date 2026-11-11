package entity

import "testing"

func FuzzDecodeProxyPrimitives(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 8))
	f.Add([]byte("ACAD_PROXY_GRAPHICS"))
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		proxy := NewAcadProxyEntity()
		proxy.GraphicsData = append([]byte(nil), data...)
		_, _ = DecodeProxyPrimitives(proxy)
	})
}
