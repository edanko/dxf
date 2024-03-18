// Package dxf is a DXF(Drawing Exchange Format) library for golang.
// ACAD2000(AC1015), ASCII format is only supported.
// http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm
package dxf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/edanko/dxf/drawing"
)

// NewDrawing creates a drawing.
func NewDrawing() (*drawing.Drawing, error) {
	return drawing.New()
}

// Create drawing from file
func FromFile(fn string) (*drawing.Drawing, error) {
	var err error
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromReader(f)
}

// Create drawing from string
func FromStringData(d string) (*drawing.Drawing, error) {
	sr := strings.NewReader(d)
	return FromReader(sr)
}

// Main logic to create a drawing
func FromReader(r io.Reader) (*drawing.Drawing, error) {
	scanner := bufio.NewScanner(r)
	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}
	var code, value string
	parsers := []func(*drawing.Drawing, int, [][2]string) error{
		ParseHeader,
		ParseClasses,
		ParseTables,
		ParseBlocks,
		ParseEntities,
		ParseObjects,
	}
	data := make([][2]string, 0)
	setparser := false
	var parser func(*drawing.Drawing, int, [][2]string) error
	line := 0
	startline := 0
	for scanner.Scan() {
		line++
		if line%2 == 1 {
			code = strings.TrimSpace(scanner.Text())
		} else {
			value = scanner.Text()
			if setparser {
				if code != "2" {
					return d, fmt.Errorf("line %d: invalid group code: %s", line, code)
				}
				ind := drawing.SectionTypeValue(strings.ToUpper(value))
				if ind < 0 {
					return d, fmt.Errorf("line %d: unknown section name: %s", line, value)
				}
				parser = parsers[ind]
				startline = line + 1
				setparser = false
			} else {
				if code == "0" {
					switch strings.ToUpper(value) {
					case "EOF":
						return d, nil
					case "SECTION":
						setparser = true
					case "ENDSEC":
						err := parser(d, startline, data)
						if err != nil {
							return d, err
						}
						data = make([][2]string, 0)
						startline = line + 1
					default:
						data = append(data, [2]string{code, scanner.Text()})
					}
				} else {
					data = append(data, [2]string{code, scanner.Text()})
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	if len(data) > 0 {
		err := parser(d, startline, data)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}
