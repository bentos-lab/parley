//go:build !windows
// +build !windows

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// installSelf replaces the existing Unix binary with the downloaded update.
// Parameters: writer receives progress messages, executablePath points to the running binary,
// binaryPath is the freshly extracted file, tmpDir holds temporary artifacts.
// Returns: an error when writing or renaming the binary fails.
func installSelf(writer io.Writer, executablePath, binaryPath, tmpDir string) error {
	_ = tmpDir
	targetDir := filepath.Dir(executablePath)
	fmt.Fprintf(writer, "Staging new binary in %s\n", targetDir)
	tempFile, err := os.CreateTemp(targetDir, "parley-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	if err := copyFile(binaryPath, tempPath); err != nil {
		return err
	}
	if err := os.Chmod(tempPath, 0o755); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tempPath, executablePath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	fmt.Fprintf(writer, "Replaced %s\n", executablePath)
	return nil
}

// copyFile copies the contents from src to dst, overwriting dst if it already exists.
// Parameters: src is the source path, dst is the already created destination file path.
// Returns: an error if the copy or flush fails.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
