package cli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const updateRepo = "bentos-lab/parley"
const unixBinaryName = "parley"
const windowsBinaryName = "parley.exe"

type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// UpdateSelf downloads the latest release archives for the current platform and installs the binary.
// Parameters: ctx controls network cancellation, writer receives progress logging.
// Returns: an error if the download or installation fails.
func UpdateSelf(ctx context.Context, writer io.Writer) error {
	osName, arch, err := normalizePlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "Detected platform %s/%s\n", osName, arch)

	version, err := fetchLatestRelease(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, "Latest version: %s\n", version)

	fileName := fmt.Sprintf("%s-%s-%s-%s", unixBinaryName, version, osName, arch)
	ext := archiveExtension(osName)
	archiveName := fileName + ext
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", updateRepo, version, archiveName)
	fmt.Fprintf(writer, "Downloading %s...\n", archiveName)

	tmpDir, err := os.MkdirTemp("", "parley-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	if err := downloadFile(ctx, url, archivePath); err != nil {
		return err
	}

	var extractedPath string
	if osName == "windows" {
		extractedPath, err = extractZip(archivePath, tmpDir)
	} else {
		extractedPath, err = extractTarGz(archivePath, tmpDir)
	}
	if err != nil {
		return err
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve current executable: %w", err)
	}
	fmt.Fprintf(writer, "Installing into %s\n", executablePath)

	if err := installSelf(writer, executablePath, extractedPath, tmpDir); err != nil {
		return err
	}
	fmt.Fprintf(writer, "Update scheduled. Restart Parley to run the new binary.\n")
	return nil
}

// normalizePlatform converts Go OS/arch names to the release naming scheme.
// Parameters: goos and goarch come from runtime.GOOS and runtime.GOARCH.
// Returns: normalized names suitable for release asset names or an error if unsupported.
func normalizePlatform(goos, goarch string) (string, string, error) {
	var osName string
	switch goos {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	case "windows":
		osName = "windows"
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", goos)
	}

	var arch string
	switch goarch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", goarch)
	}
	return osName, arch, nil
}

// fetchLatestRelease retrieves the tag name of the latest release from GitHub.
// Parameters: ctx controls the HTTP request lifecycle.
// Returns: the release tag or an error if the request fails.
func fetchLatestRelease(ctx context.Context) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", updateRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected GitHub status %d", resp.StatusCode)
	}
	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("empty release tag")
	}
	return release.TagName, nil
}

// archiveExtension chooses the archive extension based on the normalized OS name.
// Parameters: osName is the normalized release OS identifier.
// Returns: the file extension (tar.gz for Unix, zip for Windows).
func archiveExtension(osName string) string {
	if osName == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// downloadFile downloads the file from the provided URL into the destination path.
// Parameters: ctx controls cancellation, url points to the remote archive, destPath is the local target.
// Returns: an error if the download or filesystem operations fail.
func downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

// extractTarGz extracts the Unix tarball and returns the path of the unpacked binary.
// Parameters: archivePath is the downloaded .tar.gz path, destDir is the temporary extraction directory.
// Returns: the binary path or an error if extraction fails.
func extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != unixBinaryName {
			continue
		}
		targetPath := filepath.Join(destDir, filepath.Base(hdr.Name))
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(hdr.Mode))
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return "", err
		}
		outFile.Close()
		return targetPath, nil
	}
	return "", fmt.Errorf("binary %s not found in archive", unixBinaryName)
}

// extractZip extracts the Windows zip archive and returns the new binary path.
// Parameters: archivePath is the downloaded .zip path, destDir is the temporary extraction directory.
// Returns: the binary path or an error if extraction fails.
func extractZip(archivePath, destDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if filepath.Base(file.Name) != windowsBinaryName {
			continue
		}
		targetPath := filepath.Join(destDir, filepath.Base(file.Name))
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", err
		}
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", err
		}
		rc.Close()
		outFile.Close()
		return targetPath, nil
	}
	return "", fmt.Errorf("binary %s not found in archive", windowsBinaryName)
}
