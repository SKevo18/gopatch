package gopatch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

var newLineChar = getNewlineCharacter()

// Represents a slice of lines in a file
type FileLines []string

// Loads a file path as a slice of strings
func (fl *FileLines) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	*fl = strings.Split(string(data), newLineChar)

	return nil
}

// Writes the slice of strings to a file
func (fl *FileLines) WriteFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	flLen := len(*fl)
	for i, packedLines := range *fl {
		unpackedLines := strings.Split(packedLines, newLineChar)
		unpackedLen := len(unpackedLines)

		for j, line := range unpackedLines {
			if line == deletedFlag {
				continue
			}

			// trim leading backslash
			if line != "" && line[0] == '\\' {
				line = line[1:]
			}

			// append new line, if it's not the last line, or last line of the last packed line
			if i != flLen-1 || (unpackedLen > 1 && j != unpackedLen-1) {
				line += newLineChar
			}
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		}
	}

	return writer.Flush()
}

type PatchLine struct {
	// Path to the file to apply the patch to
	FilePath string

	// Line number to apply the patch from
	LineFrom int
	// Apply patch until this line number
	LineTo int

	// Content to replace the line with
	//
	// If multiple strings are provided in the slice,
	// the content will be merged according to the
	// specified merge strategy.
	Content []string

	// If `true`, the content will overwrite the line
	// If `false`, the original lines will be moved down and the content will be inserted
	Overwrite bool
}

func (pl *PatchLine) parseHeader(line string) error {
	var action rune

	// get tokens
	_, err := fmt.Sscanf(
		line, "@ %c %s %d %d %t",
		&action, &pl.FilePath,
		&pl.LineFrom, &pl.LineTo,
		&pl.Overwrite,
	)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("invalid patch header: `%s`", line)
		}
		return err
	}

	// determine action
	switch action {
	case '+':
		break
	case '-':
		pl.Content = nil
	default:
		return fmt.Errorf("unknown action: %c", action)
	}

	return nil
}

func (pl *PatchLine) String() string {
	action := "-"
	content := strings.Join(pl.Content, newLineChar)

	if len(pl.Content) > 0 {
		action = "+"
		content += newLineChar
	}

	header := fmt.Sprintf("@ %s %s %d %d %t", action, pl.FilePath, pl.LineFrom, pl.LineTo, pl.Overwrite)
	line := header + newLineChar + content

	return line
}

func getNewlineCharacter() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
