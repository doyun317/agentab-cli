package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultReleaseURL = "https://api.github.com/repos/pinchtab/pinchtab/releases/latest"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type InstallResult struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type Installer struct {
	Home       string
	Client     *http.Client
	ReleaseURL string
}

func New(home string) *Installer {
	return &Installer{
		Home:       home,
		Client:     &http.Client{Timeout: 30 * time.Second},
		ReleaseURL: defaultReleaseURL,
	}
}

func (i *Installer) ManagedBinaryPath() string {
	name := "pinchtab"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(i.Home, "bin", name)
}

func (i *Installer) ResolveBinaryPath() (string, bool, error) {
	if custom := os.Getenv("AGENTAB_PINCHTAB_BIN"); custom != "" {
		if _, err := os.Stat(custom); err != nil {
			return "", false, err
		}
		return custom, false, nil
	}
	if path, err := exec.LookPath("pinchtab"); err == nil {
		return path, false, nil
	}
	managed := i.ManagedBinaryPath()
	if _, err := os.Stat(managed); err == nil {
		return managed, true, nil
	}
	return "", false, os.ErrNotExist
}

func (i *Installer) EnsureInstalled(ctx context.Context) (InstallResult, error) {
	path, _, err := i.ResolveBinaryPath()
	if err == nil {
		return InstallResult{Path: path}, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return InstallResult{}, err
	}
	if os.Getenv("AGENTAB_SKIP_INSTALL") == "1" {
		return InstallResult{}, fmt.Errorf("pinchtab not found and installation skipped")
	}
	return i.InstallManaged(ctx)
}

func (i *Installer) InstallManaged(ctx context.Context) (InstallResult, error) {
	release, err := i.fetchRelease(ctx)
	if err != nil {
		return InstallResult{}, err
	}

	assetName, err := assetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return InstallResult{}, err
	}
	asset, err := findAsset(release.Assets, assetName)
	if err != nil {
		return InstallResult{}, err
	}

	checksum, err := assetChecksum(ctx, i.Client, asset, release.Assets)
	if err != nil {
		return InstallResult{}, err
	}
	data, err := downloadBytes(ctx, i.Client, asset.BrowserDownloadURL)
	if err != nil {
		return InstallResult{}, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != checksum {
		return InstallResult{}, fmt.Errorf("pinchtab checksum mismatch for %s", asset.Name)
	}

	path := i.ManagedBinaryPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return InstallResult{}, err
	}
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o644
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		return InstallResult{}, err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o755); err != nil {
			return InstallResult{}, err
		}
	}

	return InstallResult{Path: path, Version: release.TagName}, nil
}

func (i *Installer) fetchRelease(ctx context.Context) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.ReleaseURL, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "agentab")

	resp, err := i.Client.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("fetch pinchtab release: status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, err
	}
	return release, nil
}

func assetName(goos, goarch string) (string, error) {
	var arch string
	switch goarch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture %s", goarch)
	}

	switch goos {
	case "darwin":
		return "pinchtab-darwin-" + arch, nil
	case "linux":
		return "pinchtab-linux-" + arch, nil
	case "windows":
		return "pinchtab-windows-" + arch + ".exe", nil
	default:
		return "", fmt.Errorf("unsupported os %s", goos)
	}
}

func findAsset(assets []githubAsset, name string) (githubAsset, error) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, nil
		}
	}
	return githubAsset{}, fmt.Errorf("release asset %s not found", name)
}

func assetChecksum(ctx context.Context, client *http.Client, asset githubAsset, assets []githubAsset) (string, error) {
	if asset.Digest != "" {
		return strings.TrimPrefix(asset.Digest, "sha256:"), nil
	}

	checksums, err := findAsset(assets, "checksums.txt")
	if err != nil {
		return "", err
	}
	data, err := downloadBytes(ctx, client, checksums.BrowserDownloadURL)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == asset.Name {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found", asset.Name)
}

func downloadBytes(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agentab")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
