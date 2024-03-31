// Reads and parses patch files.
package gopatch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const deletedFlag = "@[DeLeTeD]@" // what are the odds, right?

// Reads a patch file and returns a slice of `PatchLine` structs.
func ReadPatchFile(patchFilePath string) ([]PatchLine, error) {
	file, err := os.Open(patchFilePath)
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
			patchLine.Content = append(patchLine.Content, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	patchLines = append(patchLines, patchLine)

	return patchLines, nil
}

// Joins multiple patch files together.
func ReadPatchFiles(patchFilePaths []string) ([]PatchLine, error) {
	var patchLines []PatchLine
	for _, patchFilePath := range patchFilePaths {
		lines, err := ReadPatchFile(patchFilePath)
		if err != nil {
			return nil, err
		}
		patchLines = append(patchLines, lines...)
	}

	return patchLines, nil
}

// Writes a slice of `PatchLine` structs as a patch file.
func WritePatchFile(patchFilePath string, patchLines []PatchLine) error {
	file, err := os.Create(patchFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i, patchLine := range patchLines {
		toWrite := patchLine.String()
		if i != len(patchLines)-1 {
			toWrite += newLineChar
		}

		if _, err := file.WriteString(toWrite); err != nil {
			return err
		}
	}

	return nil
}

// Patches a file with a patch file.
// Note: only the patches matching header filename are applied.
func PatchFile(filePath string, outputPath string, patchLines []PatchLine) error {
	patchedLines, err := getPatchedLines(filePath, patchLines)
	if err != nil {
		return err
	}

	if err := patchedLines.WriteFile(outputPath); err != nil {
		return err
	}

	return nil
}

// Patches a directory with a patch file.
func PatchDir(dirPath string, outputDir string, patchLines []PatchLine) error {
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// determine paths
		relPath, _ := filepath.Rel(dirPath, path)
		outputPath := filepath.Join(outputDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return err
		}

		// find relevant patch lines
		relevantPatchLines := []PatchLine{}
		for _, patchLine := range patchLines {
			if patchLine.FilePath != relPath {
				continue
			}
			relevantPatchLines = append(relevantPatchLines, patchLine)
		}
		if len(relevantPatchLines) == 0 {
			// no patches for this file
			if err := copyFile(path, outputPath); err != nil {
				return err
			}
			return nil
		}

		// get original lines and patch them
		fileLines := FileLines{}
		if err := fileLines.LoadFile(path); err != nil {
			return err
		}
		for _, patchLine := range relevantPatchLines {
			if patchLine.FilePath != relPath {
				continue
			}

			if err := fileLines.applyPatch(patchLine); err != nil {
				return err
			}
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

// Copies a file from source to destination.
func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

// Returns patched lines of a file.
func getPatchedLines(filePath string, patchLines []PatchLine) (FileLines, error) {
	fileLines := FileLines{}
	if err := fileLines.LoadFile(filePath); err != nil {
		return nil, err
	}

	for _, patchLine := range patchLines {
		if !strings.HasSuffix(patchLine.FilePath, filePath) {
			continue
		}

		if err := fileLines.applyPatch(patchLine); err != nil {
			return nil, err
		}
	}

	return fileLines, nil
}

// Applies a patch to a slice of lines.
func (fl *FileLines) applyPatch(patchLine PatchLine) error {
	if patchLine.LineFrom <= 0 || patchLine.LineFrom > len(*fl) {
		return fmt.Errorf("line number out of range (from): %d (%s)", patchLine.LineFrom, patchLine.FilePath)
	}
	if patchLine.LineTo != 0 && (patchLine.LineTo < patchLine.LineFrom || patchLine.LineTo > len(*fl)) {
		return fmt.Errorf("\"to\" line range out of bounds, or before \"from\": %d (%s)", patchLine.LineTo, patchLine.FilePath)
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
