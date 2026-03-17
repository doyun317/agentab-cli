package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentab/agentab-cli/internal/state"
)

func TestDefaultArtifactPathUsesArtifactsDir(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	got := defaultArtifactPath(
		store,
		"demo session",
		"tab_1",
		"screenshot",
		"jpg",
		time.Date(2026, 3, 17, 6, 0, 0, 0, time.UTC),
	)

	wantSuffix := filepath.Join("artifacts", "demo_session", "tab_1", "20260317T060000Z-screenshot.jpg")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("defaultArtifactPath() = %q, want suffix %q", got, wantSuffix)
	}
}

func TestSaveSnapshotArtifactWritesJSONFile(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	meta, err := saveSnapshotArtifact(
		store,
		"demo",
		"tab_1",
		"",
		map[string]any{"nodes": []map[string]any{{"ref": "e1"}}},
		time.Date(2026, 3, 17, 6, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("saveSnapshotArtifact() error = %v", err)
	}

	path, _ := meta["path"].(string)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(raw), "\"nodes\"") {
		t.Fatalf("snapshot artifact = %q, want json content", string(raw))
	}
}

func TestSaveBinaryArtifactWritesJPEGFile(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	meta, err := saveBinaryArtifact(
		store,
		"demo",
		"tab_1",
		"screenshot",
		"",
		"jpg",
		"image/jpeg",
		[]byte("jpeg-bytes"),
		time.Date(2026, 3, 17, 6, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("saveBinaryArtifact() error = %v", err)
	}

	path, _ := meta["path"].(string)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(raw) != "jpeg-bytes" {
		t.Fatalf("binary artifact = %q, want jpeg-bytes", string(raw))
	}
}
