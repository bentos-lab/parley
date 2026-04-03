package llmamodel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureModelDownloadsHTTP(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cp := []byte("model-bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(cp)
	}))
	defer server.Close()
	modelURL := server.URL + "/test.gguf"
	ctx := context.Background()
	path, err := EnsureModel(ctx, modelURL)
	require.NoError(t, err)
	expected := filepath.Join(getHomeDir(t), modelsDir, hashModelURL(modelURL)+".gguf")
	require.Equal(t, expected, path)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, cp, data)
}

func TestEnsureModelCopiesFileScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	src := t.TempDir()
	filePath := filepath.Join(src, "local.gguf")
	require.NoError(t, os.WriteFile(filePath, []byte("local"), 0o644))
	modelURL := "file:://" + filePath
	path, err := EnsureModel(context.Background(), modelURL)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(getHomeDir(t), modelsDir, hashModelURL(normalizeFileScheme(modelURL))+".gguf"), path)
	require.FileExists(t, path)
}

func TestEnsureModelSingleflightDedupesDownloads(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var mu sync.Mutex
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests++
		mu.Unlock()
		w.Write([]byte("payload"))
	}))
	defer server.Close()
	modelURL := server.URL + "/dedupe.gguf"
	ctx := context.Background()
	var wg sync.WaitGroup
	paths := make([]string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			path, err := EnsureModel(ctx, modelURL)
			require.NoError(t, err)
			paths[index] = path
		}(i)
	}
	wg.Wait()
	require.Equal(t, 1, requests)
	require.Equal(t, paths[0], paths[1])
}

func getHomeDir(t *testing.T) string {
	t.Helper()
	dir, err := os.UserHomeDir()
	require.NoError(t, err)
	return dir
}
