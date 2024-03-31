package gopatch_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SKevo18/gopatch"
	"github.com/stretchr/testify/assert"
)

const (
	testPatchFilePath = "fixtures/test.gopatch"
	testPatchFileNoCommentsPath = "fixtures/test_nocomment.gopatch"
	testInputDir      = "fixtures/original"
	textWantDir       = "fixtures/want"
)

func TestPatchDir(t *testing.T) {
	tempOutputDir := t.TempDir()

	patchLines, err := gopatch.ReadPatchFiles([]string{testPatchFilePath, testPatchFileNoCommentsPath})
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

func TestWritePatchFile(t *testing.T) {
	tempOutputDir := t.TempDir()
	tempPatchFile := filepath.Join(tempOutputDir, "test.gopatch")

	patchLines, err := gopatch.ReadPatchFile(testPatchFileNoCommentsPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := gopatch.WritePatchFile(tempPatchFile, patchLines); err != nil {
		t.Fatal(err)
	}

	if err := compareFiles(t, tempPatchFile, testPatchFileNoCommentsPath); err != nil {
		t.Fatal(err)
	}
}

func compareDirs(t *testing.T, have string, want string) error {
	return filepath.Walk(want, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(want, path)
		if err != nil {
			return err
		}

		havePath := filepath.Join(have, relPath)
		haveInfo, err := os.Stat(havePath)
		if err != nil {
			return fmt.Errorf("wanted `%s` is missing", havePath)
		}

		if info.IsDir() {
			if haveInfo.IsDir() {
				return nil
			}
			return fmt.Errorf("want %s is not a directory", havePath)
		}

		if !haveInfo.IsDir() {
			if err := compareFiles(t, path, havePath); err != nil {
				return err
			}
		}

		return nil
	})
}

func compareFiles(t *testing.T, havePath string, wantPath string) error {
	haveData, err := os.ReadFile(havePath)
	if err != nil {
		return err
	}
	wantData, err := os.ReadFile(wantPath)
	if err != nil {
		return err
	}

	if !assert.Equal(t, string(wantData), string(haveData)) {
		return fmt.Errorf("files %s do not match", havePath)
	}

	return nil
}
