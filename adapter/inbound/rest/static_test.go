package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// TestRegisterStaticAssets ensures the embedded SPA files are served for / and asset paths.
func TestRegisterStaticAssets(t *testing.T) {
	router := chi.NewRouter()
	RegisterStaticAssets(router)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Contains(t, string(data), "<div id=\"root\">")

	resp, err = http.Get(server.URL + "/logo.png")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	_, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
}
