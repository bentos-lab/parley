package install

import (
	"os"
	"os/exec"
	"runtime"
)

// Run executes the provided shell command as part of an installation step.
// Parameters: command is the platform-specific installer command to run.
// Returns: an error if the command fails to execute.
func Run(command string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	default:
		cmd := exec.Command("sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
}
