package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "linux", goarch: "amd64", want: "pinchtab-linux-amd64"},
		{goos: "linux", goarch: "arm64", want: "pinchtab-linux-arm64"},
		{goos: "darwin", goarch: "amd64", want: "pinchtab-darwin-amd64"},
		{goos: "windows", goarch: "arm64", want: "pinchtab-windows-arm64.exe"},
	}
	for _, tt := range tests {
		got, err := assetName(tt.goos, tt.goarch)
		if err != nil {
			t.Fatalf("assetName(%q, %q) error = %v", tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Fatalf("assetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestInstallManaged(t *testing.T) {
	asset, err := assetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported platform under test: %v", err)
	}

	payload := []byte("pinchtab-binary")
	sum := sha256.Sum256(payload)
	release := githubRelease{
		TagName: "v0.0.1-test",
		Assets: []githubAsset{
			{
				Name:               asset,
				BrowserDownloadURL: "/downloads/" + asset,
				Digest:             "sha256:" + hex.EncodeToString(sum[:]),
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/release":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(release)
		case "/downloads/" + asset:
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	release.Assets[0].BrowserDownloadURL = server.URL + "/downloads/" + asset

	inst := New(t.TempDir())
	inst.ReleaseURL = server.URL + "/release"
	inst.Client = server.Client()

	result, err := inst.InstallManaged(context.Background())
	if err != nil {
		t.Fatalf("InstallManaged() error = %v", err)
	}
	if result.Version != "v0.0.1-test" {
		t.Fatalf("InstallManaged().Version = %q, want v0.0.1-test", result.Version)
	}
	if filepath.Base(result.Path) == "" {
		t.Fatalf("InstallManaged().Path is empty")
	}
}
