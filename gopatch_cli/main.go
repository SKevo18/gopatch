package main

import (
	"fmt"
	"os"

	"github.com/SKevo18/gopatch"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("gopatch v0.0.1 (https://github.com/SKevo18/gopatch)")
		fmt.Printf("Usage: %s <original-dir> <output-dir> <patch-file>\n", os.Args[0])
		os.Exit(1)
	}
	originalDir := os.Args[1]
	outputDir := os.Args[2]
	patchFile := os.Args[3]

	patchLines, err := gopatch.ReadPatchFile(patchFile)
	if err != nil {
		fmt.Printf("Failed to read patch file %s: %v", patchFile, err)
		os.Exit(1)
	}
	if err := gopatch.PatchDir(originalDir, outputDir, patchLines); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
