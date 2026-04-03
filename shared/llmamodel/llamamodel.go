package llmamodel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/singleflight"
)

const (
	modelsDir = ".bentos/parley/models"
)

var ensureGroup singleflight.Group

// EnsureModel downloads or copies a GGUF model from the provided URL into the
// local cache and returns the cached path. It deduplicates concurrent work
// for the same URL and respects context cancellations.
func EnsureModel(ctx context.Context, modelURL string) (string, error) {
	urlStr := strings.TrimSpace(modelURL)
	if urlStr == "" {
		return "", fmt.Errorf("model_url is required")
	}
	urlStr = normalizeFileScheme(urlStr)
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("parse model_url: %w", err)
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("model_url missing scheme")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cacheDir, err := modelCacheDir()
	if err != nil {
		return "", err
	}
	hexKey := hashModelURL(urlStr)
	ext := filepath.Ext(parsed.Path)
	cachedPath := filepath.Join(cacheDir, hexKey+ext)
	if exists(cachedPath) {
		return cachedPath, nil
	}
	result, err, _ := ensureGroup.Do(hexKey+ext, func() (any, error) {
		if exists(cachedPath) {
			return cachedPath, nil
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		switch parsed.Scheme {
		case "file":
			if err := copyFromFile(parsed, cachedPath); err != nil {
				return "", err
			}
		case "http", "https":
			if err := downloadFile(ctx, urlStr, cachedPath); err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("unsupported scheme %q", parsed.Scheme)
		}
		return cachedPath, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// normalizeFileScheme rewrites the legacy `file:://` prefix to a standard `file://` scheme.
// Parameters: urlStr is the raw URL provided by the user.
// Returns: normalized URL.
func normalizeFileScheme(urlStr string) string {
	if strings.HasPrefix(urlStr, "file:://") {
		return "file://" + strings.TrimPrefix(urlStr, "file:://")
	}
	return urlStr
}

// hashModelURL calculates a deterministic hash for deduplication of cached model files.
// Parameters: modelURL is the location of the model.
// Returns: SHA-256 hex digest.
func hashModelURL(modelURL string) string {
	sum := sha256.Sum256([]byte(modelURL))
	return hex.EncodeToString(sum[:])
}

// modelCacheDir ensures the cache directory exists and returns its path.
// Parameters: none.
// Returns: the cache path and any error that occurred.
func modelCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	path := filepath.Join(home, modelsDir)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create model cache dir: %w", err)
	}
	return path, nil
}

// exists reports whether the provided path already points to a file.
// Parameters: path is the filesystem path to probe.
// Returns: true when the file exists and is not a directory.
func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// copyFromFile copies a local file referenced via file:// into the cache path.
// Parameters: parsed holds the parsed URL, dest is the destination cache file path.
// Returns: any error encountered while reading or writing.
func copyFromFile(parsed *url.URL, dest string) error {
	srcPath := filepath.FromSlash(parsed.Path)
	if srcPath == "" {
		return fmt.Errorf("file path is empty")
	}
	srcPath = filepath.Clean(srcPath)
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("read source file: %w", err)
	}
	input, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer input.Close()
	return writeFileFromStream(input, dest)
}

// downloadFile streams the remote model bytes into the destination path.
// Parameters: ctx controls cancellation, modelURL is the download target, dest is where the file is stored.
// Returns: any download or write error.
func downloadFile(ctx context.Context, modelURL, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelURL, nil)
	if err != nil {
		return fmt.Errorf("prepare download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download model: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s", strings.TrimSpace(string(body)))
	}
	return writeFileFromStream(resp.Body, dest)
}

// writeFileFromStream writes the streamed bytes to a temp file and renames it atomically.
// Parameters: reader streams the source bytes, dest is the final cache file.
// Returns: any error encountered during writing or renaming.
func writeFileFromStream(reader io.Reader, dest string) error {
	tmp := dest + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("finalize cache file: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("rename cache file: %w", err)
	}
	return nil
}
