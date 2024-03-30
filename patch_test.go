package gopatch_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/SKevo18/gopatch"
)

const (
	testPatchFilePath = "fixtures/test.gopatch"
	testInputDir      = "fixtures/original"
	textWantDir       = "fixtures/want"
)

func TestPatchDir(t *testing.T) {
	tempOutputDir := t.TempDir()

	patchLines, err := gopatch.ReadPatchFile(testPatchFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := gopatch.PatchDir(testInputDir, tempOutputDir, patchLines); err != nil {
		t.Fatal(err)
	}

	if err := compareDirs(t, tempOutputDir, textWantDir); err != nil {
		t.Fatal(err)
	}
}

func compareDirs(t *testing.T, have string, want string) error {
	return filepath.Walk(have, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(have, path)
		if err != nil {
			return err
		}

		wantPath := filepath.Join(want, relPath)
		wantInfo, err := os.Stat(wantPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if wantInfo.IsDir() {
				return nil
			}
			return fmt.Errorf("want %s is not a directory", wantPath)
		}

		if !wantInfo.IsDir() {
			haveData, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			wantData, err := os.ReadFile(wantPath)
			if err != nil {
				return err
			}

			if !assert.Equal(t, string(haveData), string(wantData)) {
				return fmt.Errorf("file %s does not match", relPath)
			}
		}

		return nil
	})
}
