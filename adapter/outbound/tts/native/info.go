package native

import "runtime"

// Info describes the native TTS tooling for the current platform.
type Info struct {
	Label          string
	Executable     string
	InstallCommand string
	InstallLink    string
	ReadyMessage   string
}

// CurrentInfo returns the native TTS info for the current OS.
func CurrentInfo() Info {
	switch runtime.GOOS {
	case "darwin":
		return Info{
			Label:        "native (say)",
			Executable:   "say",
			ReadyMessage: "macOS includes the `say` command by default. No install needed.",
		}
	case "windows":
		return Info{
			Label:        "native (windows speech api)",
			Executable:   "powershell",
			InstallLink:  "https://learn.microsoft.com/en-us/dotnet/api/system.speech.synthesis.speechsynthesizer",
			ReadyMessage: "Windows Speech API is built in. No install needed.",
		}
	default:
		return Info{
			Label:          "native (espeak)",
			Executable:     "espeak",
			InstallCommand: "sudo apt-get update && sudo apt-get install -y espeak",
			InstallLink:    "https://github.com/espeak-ng/espeak-ng",
		}
	}
}
