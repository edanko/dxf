package patterns

import "testing"

func FuzzParsePAT(f *testing.F) {
	f.Add([]byte("*DASH, dashes\n45, 0,0, 1,1, 0.5,-0.5\n"))
	f.Add([]byte("; comment only\n"))
	f.Add([]byte("*BROKEN\nno,valid,pattern\n"))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		if got := ParsePAT(data, "fuzz.pat"); got == nil {
			t.Fatal("ParsePAT returned nil")
		}
	})
}
