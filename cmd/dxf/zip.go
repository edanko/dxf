package main

import (
	"flag"
	"fmt"

	"github.com/edanko/dxf/filemanagement"
)

func cmdZip(args []string) error {
	fs, err := newFlagSet("zip", args, func(*flag.FlagSet) {})
	if err != nil || fs == nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return usageError("usage: dxf zip <archive.zip> <file> [file...]")
	}
	archive := rest[0]
	files := rest[1:]
	if err := filemanagement.CreateZipFromFiles(archive, files); err != nil {
		return dataError("%v", err)
	}
	fmt.Printf("created %s with %d entries\n", archive, len(files))
	return nil
}

func cmdUnzip(args []string) error {
	fs, err := newFlagSet("unzip", args, func(*flag.FlagSet) {})
	if err != nil || fs == nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return usageError("usage: dxf unzip <archive.zip> [target-dir]")
	}
	archive := rest[0]
	target := "."
	if len(rest) > 1 {
		target = rest[1]
	}
	extracted, err := filemanagement.ExtractZip(archive, target)
	if err != nil {
		return dataError("%v", err)
	}
	fmt.Printf("extracted %d entries to %s\n", len(extracted), target)
	return nil
}
