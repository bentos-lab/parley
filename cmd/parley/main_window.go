//go:build windows
// +build windows

package main

import "github.com/spf13/cobra"

func init() {
	// Disable hook for double click on windows
	cobra.MousetrapHelpText = ""
}
