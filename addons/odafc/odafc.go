// Package odafc integrates with the ODA File Converter command line tool.
package odafc

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/addons"
)

func init() { _ = addons.Register("odafc", capability{}) }

type capability struct{}

func (capability) Name() string        { return "odafc" }
func (capability) Description() string { return "ODA File Converter integration" }

// CommandRunner invokes an external command (injectable for tests).
type CommandRunner interface {
	Run(name string, args []string, stdin []byte) (stdout, stderr []byte, err error)
}

// ODAConverter converts files via the ODA File Converter.
type ODAConverter struct {
	Runner CommandRunner
}

// New creates a converter backed by runner.
func New(runner CommandRunner) *ODAConverter { return &ODAConverter{Runner: runner} }

// Convert runs the converter; a missing tool yields an actionable error.
func (c *ODAConverter) Convert(args []string, stdin []byte) (stdout, stderr []byte, err error) {
	if c == nil || c.Runner == nil {
		return nil, nil, fmt.Errorf("ODA File Converter not available: install the ODA File Converter or provide a CommandRunner")
	}
	out, errOut, err := c.Runner.Run("ODAFileConverter", args, stdin)
	if err != nil {
		return out, errOut, fmt.Errorf("oda file converter: %w: %s", err, strings.TrimSpace(string(errOut)))
	}
	return out, errOut, nil
}
