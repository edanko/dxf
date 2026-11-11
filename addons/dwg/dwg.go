package dwg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/edanko/dxf/addons"
)

// CommandRunner executes an external command.
type CommandRunner interface {
	Run(name string, args []string, stdin []byte) (stdout, stderr []byte, err error)
}

// Converter delegates DWG conversion to an external ODA File Converter.
type Converter struct {
	Runner     CommandRunner
	Executable string
}

// New creates a DWG converter.
func New(runner CommandRunner, executable string) *Converter {
	return &Converter{Runner: runner, Executable: executable}
}

// Convert converts inputPath (DWG) to outputPath (DXF) through the configured
// external converter. The final output is written atomically: a failed
// conversion never leaves a partial output file behind.
func (c *Converter) Convert(inputPath, outputPath string) error {
	if c == nil || c.Runner == nil || c.Executable == "" {
		return fmt.Errorf("DWG converter not available: configure an ODA File Converter executable and CommandRunner")
	}

	dirname := filepath.Dir(outputPath)
	tempPath := filepath.Join(dirname, "."+filepath.Base(outputPath)+".tmp")
	_ = os.Remove(tempPath)

	_, stderr, err := c.Runner.Run(c.Executable, []string{inputPath, tempPath}, nil)
	if err != nil {
		_ = os.Remove(tempPath)
		if len(stderr) > 0 {
			return fmt.Errorf("DWG conversion failed: %w: %s", err, stderr)
		}
		return fmt.Errorf("DWG conversion failed: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("DWG conversion failed: cannot write %s: %w", outputPath, err)
	}
	return nil
}

type capability struct{}

func (capability) Name() string { return "dwg" }

func (capability) Description() string {
	return "DWG to DXF conversion through an external ODA File Converter"
}

func init() {
	_ = addons.Register("dwg", capability{})
}
