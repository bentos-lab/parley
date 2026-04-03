package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configData := `
[llm.openai]
base_url = "https://example.com"
api_key = "file-key"
model = "file-model"

[tts.inworld]
api_key = "file-inworld-key"
model = "inworld-tts-1.5-mini"
`
	if err := os.WriteFile(path, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("OPENAI_MODEL", "env-model")
	t.Setenv("INWORLD_MODEL", "env-inworld-model")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.OpenAI.BaseURL != "https://example.com" {
		t.Fatalf("expected base url from file, got %q", cfg.OpenAI.BaseURL)
	}
	if cfg.OpenAI.Model != "env-model" {
		t.Fatalf("expected OpenAI model override, got %q", cfg.OpenAI.Model)
	}
	if cfg.InworldModel != "env-inworld-model" {
		t.Fatalf("expected Inworld model override, got %q", cfg.InworldModel)
	}
	if cfg.OpenAI.APIKey != "file-key" || cfg.InworldAPIKey != "file-inworld-key" {
		t.Fatalf("expected API keys from file, got %q / %q", cfg.OpenAI.APIKey, cfg.InworldAPIKey)
	}
}

func TestLoadConfigProviderDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configData := `
[llm.openai]
api_key = "file-key"

[tts.inworld]
api_key = "file-inworld-key"
`
	if err := os.WriteFile(path, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.LLMProvider != "openai" {
		t.Fatalf("expected default llm provider, got %q", cfg.LLMProvider)
	}
	if cfg.TTSProvider != "native" {
		t.Fatalf("expected default tts provider, got %q", cfg.TTSProvider)
	}
}

func TestLoadConfigInvalidProviders(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configData := `
[llm]
provider = "unknown"

[tts]
provider = "nope"
`
	if err := os.WriteFile(path, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid providers")
	}
}

func TestSetNestedValueWritesConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := WriteFileMap(map[string]any{}); err != nil {
		t.Fatalf("write config map: %v", err)
	}
	cfgMap, err := ReadFileMap()
	if err != nil {
		t.Fatalf("read config map: %v", err)
	}
	SetNestedValue(cfgMap, []string{"llm", "openai"}, "model", "test-model")
	if err := WriteFileMap(cfgMap); err != nil {
		t.Fatalf("write config map: %v", err)
	}
	reloaded, err := ReadFileMap()
	if err != nil {
		t.Fatalf("read config map: %v", err)
	}
	llmSection, ok := reloaded["llm"].(map[string]any)
	if !ok {
		t.Fatalf("expected llm section")
	}
	openaiSection, ok := llmSection["openai"].(map[string]any)
	if !ok {
		t.Fatalf("expected openai section")
	}
	if openaiSection["model"] != "test-model" {
		t.Fatalf("expected model to be updated, got %v", openaiSection["model"])
	}
}

func TestWriteFileMapStripsEmptyParentTables(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfg := map[string]any{
		"tts": map[string]any{
			"inworld": map[string]any{
				"api_key": "test-key",
				"model":   "test-model",
			},
		},
	}
	if err := WriteFileMap(cfg); err != nil {
		t.Fatalf("write config map: %v", err)
	}
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "[tts]\n") {
		t.Fatalf("did not expect empty [tts] header")
	}
	if !strings.Contains(content, "[tts.inworld]\n") {
		t.Fatalf("expected [tts.inworld] header")
	}
}
