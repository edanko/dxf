// Package appsettings resolves CLI configuration from defaults, a JSON
// config file, and DXF_ prefixed environment variables (in increasing
// precedence).
package appsettings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// Config carries the configurable defaults of the dxf CLI.
type Config struct {
	Output     string  `json:"output"`
	Backend    string  `json:"backend"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Background string  `json:"background"`
}

// Default returns the built-in defaults.
func Default() Config {
	return Config{Backend: "svg", Width: 800, Height: 600, Background: "#ffffff"}
}

// ConfigPath returns the config file path: DXF_CONFIG when set, otherwise
// ~/.config/dxf/config.json.
func ConfigPath() string {
	if p := os.Getenv("DXF_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "dxf", "config.json")
}

// Load resolves the configuration: defaults < config file < DXF_* env.
func Load() Config {
	c := Default()
	if path := ConfigPath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(data, &c)
		}
	}
	if v := os.Getenv("DXF_OUTPUT"); v != "" {
		c.Output = v
	}
	if v := os.Getenv("DXF_BACKEND"); v != "" {
		c.Backend = v
	}
	if v := os.Getenv("DXF_WIDTH"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.Width = f
		}
	}
	if v := os.Getenv("DXF_HEIGHT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.Height = f
		}
	}
	if v := os.Getenv("DXF_BACKGROUND"); v != "" {
		c.Background = v
	}
	return c
}
