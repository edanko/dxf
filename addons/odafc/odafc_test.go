package odafc_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/odafc"
)

type fakeRunner struct {
	gotName string
	gotArgs []string
	stdin   []byte
	fail    bool
}

func (f *fakeRunner) Run(name string, args []string, stdin []byte) ([]byte, []byte, error) {
	f.gotName = name
	f.gotArgs = args
	f.stdin = stdin
	if f.fail {
		return nil, []byte("boom"), errors.New("exit 1")
	}
	return []byte("converted"), nil, nil
}

func TestConvertWithRunner(t *testing.T) {
	f := &fakeRunner{}
	c := odafc.New(f)
	out, _, err := c.Convert([]string{"in.dwg", "out"}, []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "converted" || f.gotName != "ODAFileConverter" || len(f.gotArgs) != 2 {
		t.Fatalf("runner not invoked correctly: %q %v", out, f.gotArgs)
	}
}

func TestConvertMissingTool(t *testing.T) {
	c := odafc.New(nil)
	_, _, err := c.Convert([]string{"in"}, nil)
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("missing tool error: %v", err)
	}

	f := &fakeRunner{fail: true}
	c2 := odafc.New(f)
	_, _, err = c2.Convert([]string{"in"}, nil)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("runner failure must surface stderr: %v", err)
	}
}
