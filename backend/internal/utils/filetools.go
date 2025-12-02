package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// mergeChunks merges all chunk files into a single destination file
func MergeChunks(tmpDir, destPath string, totalChunks int) error {
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%d", i))
		chunk, err := os.Open(chunkPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dest, chunk); err != nil {
			chunk.Close()
			return err
		}
		chunk.Close()
	}
	return nil
}

// extractZip extracts a ZIP file to the specified destination directory
// If files already exist, they will be overwritten
func ExtractZip(zipPath, destDir string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0777); err != nil { // full read/write/exec permissions
		switch {
		case os.IsPermission(err):
			return fmt.Errorf("permission denied while creating directory: %s", destDir)
		case errors.Is(err, syscall.ENOSPC):
			return fmt.Errorf("insufficient storage space to create directory")
		case errors.Is(err, syscall.ENAMETOOLONG):
			return fmt.Errorf("path too long: %s", destDir)
		case errors.Is(err, syscall.EINVAL):
			return fmt.Errorf("invalid directory name: %s", destDir)
		default:
			return fmt.Errorf("unexpected error creating directory %s: %v", destDir, err)
		}
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0777); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fpath), 0777); err != nil {
			return err
		}

		// Create/truncate file (overwrites if exists)
		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}
