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
	for i, line := range *fl {
		// unpack lines
		lines := strings.Split(line, "\n") // TODO: do we need to strip \r?

		for _, l := range lines {
			// skip deleted lines
			if l == deletedFlag {
				continue
			}

			toWrite := l
			if i < len(*fl)-1 {
				toWrite += newLineChar
			}
			if _, err := writer.WriteString(toWrite); err != nil {
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

func getNewlineCharacter() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
