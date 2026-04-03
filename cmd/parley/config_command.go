package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/config"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [section.key=value ...]",
		Short: "View or edit the config file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ConfigPath()
			if err != nil {
				return err
			}
			cfgMap, err := config.ReadFileMap()
			if err != nil {
				return err
			}
			for _, arg := range args {
				pathParts, key, value, err := parseConfigAssignment(arg)
				if err != nil {
					return err
				}
				config.SetNestedValue(cfgMap, pathParts, key, value)
			}
			if err := config.WriteFileMap(cfgMap); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Config Updated")
			fmt.Fprintf(os.Stdout, "Updated %d value(s)\n", len(args))
			fmt.Fprintf(os.Stdout, "File: %s\n", path)
			return nil
		},
	}
	return cmd
}

func parseConfigAssignment(input string) ([]string, string, string, error) {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return nil, "", "", fmt.Errorf("invalid assignment %q (expected section.key=value)", input)
	}
	left := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	keyPath := strings.Split(left, ".")
	if len(keyPath) < 2 {
		return nil, "", "", fmt.Errorf("invalid key path %q (expected section.key)", parts[0])
	}
	pathParts := keyPath[:len(keyPath)-1]
	key := keyPath[len(keyPath)-1]
	if key == "" {
		return nil, "", "", fmt.Errorf("invalid key path %q", parts[0])
	}
	return pathParts, key, value, nil
}
