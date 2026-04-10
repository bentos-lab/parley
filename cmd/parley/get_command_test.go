package main

import (
	"context"
	"testing"

	clicore "github.com/bentos-lab/parley/adapter/inbound/cli"
	jsonoutput "github.com/bentos-lab/parley/adapter/inbound/cli/json"
	"github.com/bentos-lab/parley/adapter/inbound/cli/normal"
	"github.com/bentos-lab/parley/adapter/inbound/cli/pretty"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// TestNewGetCommandRequiresID verifies get command rejects missing IDs.
// Parameters: t provides the test context.
// Returns: nothing.
func TestNewGetCommandRequiresID(t *testing.T) {
	cmd := newGetCommand(context.Background(), nil, clicore.RuntimeInfo{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.EqualError(t, err, "id is required")
}

// TestGetOutputForCommandDefaultsToPretty verifies get format defaults to pretty.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetOutputForCommandDefaultsToPretty(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().String("format", "pretty", "output format")

	output, err := getOutputForCommand(cmd)
	require.NoError(t, err)
	require.IsType(t, &pretty.Output{}, output)
}

// TestGetOutputForCommandSupportsNormal verifies get format accepts normal.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetOutputForCommandSupportsNormal(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().String("format", "pretty", "output format")
	require.NoError(t, cmd.Flags().Set("format", "normal"))

	output, err := getOutputForCommand(cmd)
	require.NoError(t, err)
	require.IsType(t, &normal.Output{}, output)
}

// TestGetOutputForCommandSupportsJSON verifies get format accepts json.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetOutputForCommandSupportsJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().String("format", "pretty", "output format")
	require.NoError(t, cmd.Flags().Set("format", "json"))

	output, err := getOutputForCommand(cmd)
	require.NoError(t, err)
	require.IsType(t, &jsonoutput.Output{}, output)
}

// TestGetOutputForCommandRejectsInvalidFormat verifies get format rejects unknown values.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetOutputForCommandRejectsInvalidFormat(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().String("format", "pretty", "output format")
	require.NoError(t, cmd.Flags().Set("format", "yaml"))

	_, err := getOutputForCommand(cmd)
	require.EqualError(t, err, `invalid format "yaml" (expected pretty, normal, or json)`)
}
