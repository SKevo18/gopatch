// Reads and parses patch files.
package gopatch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const deletedFlag = "@[DeLeTeD]@" // what are the odds, right?

// Reads a patch file and returns a slice of `PatchLine` structs.
func readPatchFile(filePath string) ([]PatchLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patchLines []PatchLine
	scanner := bufio.NewScanner(file)

	var patchLine PatchLine
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		if line[0] == '@' {
			// we already know file path, so this is next header
			if patchLine.FilePath != "" {
				patchLines = append(patchLines, patchLine)
				patchLine = PatchLine{}
			}

			if err := patchLine.parseHeader(line); err != nil {
				return nil, err
			}
		} else {
			if line[0] == '\\' {
				line = line[1:] // trim leading backslash
			}
			patchLine.Content = append(patchLine.Content, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	patchLines = append(patchLines, patchLine)

	return patchLines, nil
}

// Patches a directory with a patch file.
func PatchDir(dirPath string, outputDir string, patchFilePath string) error {
	patchLines, err := readPatchFile(patchFilePath)
	if err != nil {
		return err
	}

	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fileLines := FileLines{}
		if err := fileLines.LoadFile(path); err != nil {
			return err
		}

		// accumulate patches
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		for _, patchLine := range patchLines {
			if patchLine.FilePath != relPath {
				continue
			}
			if err := fileLines.applyPatch(patchLine); err != nil {
				return err
			}

		}

		outputPath := filepath.Join(outputDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return err
		}

		if err := fileLines.WriteFile(outputPath); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Applies a patch to a slice of lines.
func (fl *FileLines) applyPatch(patchLine PatchLine) error {
	if patchLine.LineFrom <= 0 || patchLine.LineFrom > len(*fl) {
		return fmt.Errorf("line number out of range (from): %d (%s)", patchLine.LineFrom, patchLine.FilePath)
	}
	if patchLine.LineTo != 0 && (patchLine.LineTo < patchLine.LineFrom || patchLine.LineTo > len(*fl)) {
		return fmt.Errorf("nonsense \"to\" line range: %d (%s)", patchLine.LineTo, patchLine.FilePath)
	}
	if patchLine.LineTo == 0 {
		patchLine.LineTo = patchLine.LineFrom
	}

	if len(patchLine.Content) == 0 {
		if patchLine.Overwrite {
			// hard delete
			for i := patchLine.LineFrom - 1; i < patchLine.LineTo; i++ {
				(*fl)[i] = deletedFlag // mark lines as deleted
			}
		} else {
			// soft delete (empty line)
			for i := patchLine.LineFrom - 1; i < patchLine.LineTo; i++ {
				(*fl)[i] = ""
			}
		}
	} else {
		if patchLine.Overwrite {
			// replace lines in place
			x := 0
			for i := patchLine.LineFrom - 1; i < patchLine.LineFrom+len(patchLine.Content)-1; i++ {
				if i >= len(*fl) {
					*fl = append(*fl, patchLine.Content[x])
				} else {
					(*fl)[i] = patchLine.Content[x]
				}
				x++
			}
		} else {
			// insert lines, shift original below
			(*fl)[patchLine.LineFrom-1] = strings.Join(patchLine.Content, newLineChar) + newLineChar + (*fl)[patchLine.LineFrom-1]
		}
	}

	return nil
}
