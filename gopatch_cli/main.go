package main

import (
	"fmt"
	"os"

	"github.com/SKevo18/gopatch"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("gopatch v1.0 (https://github.com/SKevo18/gopatch)")
		fmt.Printf("Usage: %s <original-dir> <output-dir> <patch-files...>\n", os.Args[0])
		os.Exit(1)
	}
	originalDir := os.Args[1]
	outputDir := os.Args[2]
	patchFiles := os.Args[3:]

	patchLines, err := gopatch.ReadPatchFiles(patchFiles)
	if err != nil {
		fmt.Printf("Failed to read a patch file: %v", err)
		os.Exit(1)
	}
	if err := gopatch.PatchDir(originalDir, outputDir, patchLines); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
