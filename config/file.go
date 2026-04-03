package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	configDirName  = ".bentos"
	configFileName = "parley.config.toml"
)

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, configDirName, configFileName), nil
}

func loadFileConfig() (fileConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return fileConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileConfig{}, nil
		}
		return fileConfig{}, fmt.Errorf("read config: %w", err)
	}
	var cfg fileConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// ReadFileMap reads the config file into a map.
func ReadFileMap() (map[string]any, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg map[string]any
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg == nil {
		cfg = map[string]any{}
	}
	return cfg, nil
}

// WriteFileMap writes the config map back to the config file.
func WriteFileMap(cfg map[string]any) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = stripEmptyTables(data)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// stripEmptyTables removes parent table headers that only contain child tables.
// It keeps array tables ([[...]]), and preserves headers that have direct keys.
func stripEmptyTables(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return data
	}

	var out []string
	lastHeaderIndex := -1
	lastHeaderHasKey := false
	lastHeaderIsArray := false

	flushHeaderIfNeeded := func() {
		if lastHeaderIndex == -1 {
			return
		}
		if !lastHeaderIsArray && !lastHeaderHasKey {
			// Remove empty parent header.
			out = append(out[:lastHeaderIndex], out[lastHeaderIndex+1:]...)
		}
		lastHeaderIndex = -1
		lastHeaderHasKey = false
		lastHeaderIsArray = false
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}

		isArrayHeader := strings.HasPrefix(trimmed, "[[") && strings.HasSuffix(trimmed, "]]")
		isHeader := strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
		if isHeader {
			flushHeaderIfNeeded()
			out = append(out, line)
			lastHeaderIndex = len(out) - 1
			lastHeaderHasKey = false
			lastHeaderIsArray = isArrayHeader
			continue
		}

		if lastHeaderIndex != -1 && strings.Contains(trimmed, "=") {
			lastHeaderHasKey = true
		}
		out = append(out, line)
	}

	flushHeaderIfNeeded()
	return []byte(strings.Join(out, "\n"))
}

// SetNestedValue sets a nested value in the provided map.
func SetNestedValue(cfg map[string]any, path []string, key string, value string) {
	current := cfg
	for _, segment := range path {
		next, ok := current[segment].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[segment] = next
		}
		current = next
	}
	current[key] = value
}
