package main

import (
	"fmt"
	"log"
	"os"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/drawing"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: binary_test <dxf_file>")
	}

	filename := os.Args[1]

	// Test format detection
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	format, err := dxf.DetectFormat(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Format detected: %s\n", format)

	// Try to load the file
	var drawing *drawing.Drawing

	if format == "binary" {
		drawing, err = dxf.FromBinaryFile(filename)
	} else {
		drawing, err = dxf.FromFile(filename)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully loaded drawing: %s\n", filename)
	fmt.Printf("Entities: %d\n", len(drawing.Entities()))
}
