package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	jsonoutput "github.com/bentos-lab/parley/adapter/inbound/cli/json"
	"github.com/bentos-lab/parley/adapter/inbound/cli/normal"
	"github.com/bentos-lab/parley/adapter/inbound/cli/pretty"
)

type createResumeOutput interface {
	cli.CreateOutput
	cli.ResumeOutput
}

func outputForCommand(cmd *cobra.Command) (createResumeOutput, error) {
	format, err := formatFlag(cmd)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(format) {
	case "pretty":
		return pretty.New(), nil
	case "json":
		return jsonoutput.New(), nil
	default:
		return nil, fmt.Errorf("invalid format %q (expected pretty or json)", format)
	}
}

func listOutputForCommand(cmd *cobra.Command) (cli.ListOutput, error) {
	format, err := formatFlag(cmd)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(format) {
	case "pretty":
		return pretty.New(), nil
	case "json":
		return jsonoutput.New(), nil
	default:
		return nil, fmt.Errorf("invalid format %q (expected pretty or json)", format)
	}
}

func getOutputForCommand(cmd *cobra.Command) (cli.GetOutput, error) {
	format, err := formatFlag(cmd)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(format) {
	case "pretty":
		return pretty.New(), nil
	case "normal":
		return normal.New(), nil
	case "json":
		return jsonoutput.New(), nil
	default:
		return nil, fmt.Errorf("invalid format %q (expected pretty, normal, or json)", format)
	}
}

func formatFlag(cmd *cobra.Command) (string, error) {
	format := "pretty"
	if cmd != nil {
		value, err := cmd.Flags().GetString("format")
		if err != nil {
			return "", err
		}
		if value != "" {
			format = value
		}
	}
	return format, nil
}
