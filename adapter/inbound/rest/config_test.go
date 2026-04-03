package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bentos-lab/parley/adapter/outbound/tts"
	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
	"github.com/bentos-lab/parley/config"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func newTestRouter(t *testing.T, opts ...HandlerOption) http.Handler {
	t.Helper()
	handler := NewHandler(opts...)
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})
	return router
}

func TestGetConfigReturnsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response ConfigResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "openai", response.LLM.Provider)
	require.Equal(t, "https://api.openai.com/v1", response.LLM.OpenAI.BaseURL)
	require.Equal(t, "gpt-4.1-mini", response.LLM.OpenAI.Model)
	require.Equal(t, "native", response.TTS.Provider)
	require.Equal(t, "inworld-tts-1.5-max", response.TTS.Inworld.Model)
}

func TestUpdateConfigWritesLLMandTTS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	router := newTestRouter(t)

	payload := ConfigRequest{
		LLM: &LLMConfigRequest{
			Provider: stringPtr("openai"),
			OpenAI: &OpenAIConfigRequest{
				BaseURL: stringPtr("https://example.com/v1"),
				APIKey:  stringPtr("api-key"),
				Model:   stringPtr("gpt-4o"),
			},
		},
		TTS: &TTSConfigRequest{
			Provider: stringPtr("inworld"),
			Inworld: &InworldConfigRequest{
				APIKey: stringPtr("inworld-key"),
				Model:  stringPtr("inworld-tts-1.5-mini"),
			},
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/api/config", mustJSONReader(t, payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var cfgResp ConfigResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &cfgResp))
	require.Equal(t, "https://example.com/v1", cfgResp.LLM.OpenAI.BaseURL)
	require.Equal(t, "api-key", cfgResp.LLM.OpenAI.APIKey)
	require.Equal(t, "gpt-4o", cfgResp.LLM.OpenAI.Model)
	require.Equal(t, "inworld", cfgResp.TTS.Provider)
	require.Equal(t, "inworld-tts-1.5-mini", cfgResp.TTS.Inworld.Model)

	cfgMap, err := config.ReadFileMap()
	require.NoError(t, err)
	ttsSection, ok := cfgMap["tts"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "inworld", ttsSection["provider"])
	path, err := config.ConfigPath()
	require.NoError(t, err)
	require.FileExists(t, path)
}

func TestUpdateConfigNativeProviderRunsInstaller(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	lookups := 0
	lookPath := func(name string) (string, error) {
		lookups++
		if lookups == 1 {
			return "", fmt.Errorf("missing")
		}
		return "/usr/bin/" + name, nil
	}
	commands := []string{}
	runInstall := func(command string) error {
		commands = append(commands, command)
		return nil
	}

	router := newTestRouter(t, WithLookPath(lookPath), WithRunInstall(runInstall))
	payload := ConfigRequest{TTS: &TTSConfigRequest{Provider: stringPtr(tts.ProviderNative)}}

	req := httptest.NewRequest(http.MethodPut, "/api/config", mustJSONReader(t, payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, commands, 1)
	require.Equal(t, native.CurrentInfo().InstallCommand, commands[0])
}

func TestUpdateConfigNativeInstallFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	lookPath := func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}
	installCalls := 0
	runInstall := func(command string) error {
		installCalls++
		return fmt.Errorf("failed")
	}

	router := newTestRouter(t, WithLookPath(lookPath), WithRunInstall(runInstall))
	payload := ConfigRequest{TTS: &TTSConfigRequest{Provider: stringPtr(tts.ProviderNative)}}

	req := httptest.NewRequest(http.MethodPut, "/api/config", mustJSONReader(t, payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, 1, installCalls)
	path, err := config.ConfigPath()
	require.NoError(t, err)
	_, statErr := os.Stat(path)
	require.True(t, os.IsNotExist(statErr))
}

func mustJSONReader(t *testing.T, value any) *bytes.Reader {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return bytes.NewReader(data)
}

func stringPtr(value string) *string {
	return &value
}
