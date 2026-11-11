package appsettings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.Backend != "svg" || c.Width != 800 || c.Height != 600 || c.Background != "#ffffff" {
		t.Fatalf("defaults: %+v", c)
	}
}

func TestEnvPrecedence(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "none.json")
	os.Setenv("DXF_CONFIG", cfgPath)
	defer os.Unsetenv("DXF_CONFIG")
	os.Setenv("DXF_WIDTH", "1234")
	defer os.Unsetenv("DXF_WIDTH")
	os.Setenv("DXF_BACKEND", "pdf")
	defer os.Unsetenv("DXF_BACKEND")

	c := Load()
	if c.Width != 1234 || c.Backend != "pdf" {
		t.Fatalf("env overrides: %+v", c)
	}
	if c.Height != 600 {
		t.Fatalf("unset values keep defaults: %+v", c)
	}
}

func TestConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"backend":"png","width":640}`), 0644); err != nil {
		t.Fatal(err)
	}
	os.Setenv("DXF_CONFIG", cfgPath)
	defer os.Unsetenv("DXF_CONFIG")

	c := Load()
	if c.Backend != "png" || c.Width != 640 {
		t.Fatalf("file values: %+v", c)
	}
	if c.Height != 600 {
		t.Fatalf("missing file values keep defaults: %+v", c)
	}
}
