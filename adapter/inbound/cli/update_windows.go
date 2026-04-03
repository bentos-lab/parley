//go:build windows
// +build windows

package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// installSelf schedules a PowerShell helper that replaces the binary once Parley exits.
// Parameters: writer receives progress messages, executablePath is the current binary path,
// binaryPath points to the downloaded binary, tmpDir houses temporary artifacts.
// Returns: an error when scheduling the helper fails.
func installSelf(writer io.Writer, executablePath, binaryPath, tmpDir string) error {
	fmt.Fprintf(writer, "Scheduling Windows updater for %s\n", executablePath)
	scriptPath := filepath.Join(tmpDir, "parley-update.ps1")
	script := buildWindowsScript(executablePath, binaryPath, tmpDir, os.Getpid())
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return fmt.Errorf("write updater script: %w", err)
	}
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start updater script: %w", err)
	}
	fmt.Fprintf(writer, "Updater launched; the binary will replace itself once Parley exits.\n")
	return nil
}

// buildWindowsScript returns the PowerShell script that waits for the parent and moves files.
// Parameters: target is the final binary path, source is the downloaded file, tmpDir is the temp root,
// pid is the process ID of the running Parley process.
// Returns: the script contents.
func buildWindowsScript(target, source, tmpDir string, pid int) string {
	return fmt.Sprintf(`$target = %q
$source = %q
$tmp = %q
$pid = %d
while (Get-Process -Id $pid -ErrorAction SilentlyContinue) {
    Start-Sleep -Seconds 1
}
try {
    Move-Item -LiteralPath $source -Destination $target -Force
} catch {
    Write-Error "Unable to replace binary: $_"
    exit 1
}
Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
`, target, source, tmpDir, pid)
}
