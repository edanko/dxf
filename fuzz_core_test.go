package dxf

import (
	"bytes"
	"io"
	"testing"

	"github.com/edanko/dxf/lldxf"
)

func FuzzASCIITagger(f *testing.F) {
	f.Add([]byte("0\nSECTION\n2\nENTITIES\n0\nENDSEC\n0\nEOF\n"))
	f.Add([]byte("0\nLINE\n10\n1.0\n20\n2.0\n30\n3.0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		tagger := lldxf.NewASCIITagger(bytes.NewReader(data))
		for tagger.Next() {
			_ = tagger.Tag()
		}
		_ = tagger.Err()
	})
}

func FuzzLoader(f *testing.F) {
	f.Add([]byte("0\nSECTION\n2\nENTITIES\n0\nPOINT\n10\n1\n20\n2\n30\n3\n0\nENDSEC\n0\nEOF\n"))
	f.Add([]byte("0\nSECTION\n2\nTABLES\n0\nTABLE\n2\nLAYER\n0\nLAYER\n2\n0\n0\nENDTAB\n0\nENDSEC\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		d, err := FromReaderWithOptions(bytes.NewReader(data), &LoadOptions{Recover: true})
		if err != nil || d == nil {
			return
		}
		_, _ = d.WriteTo(io.Discard)
	})
}
