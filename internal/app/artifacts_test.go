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
	if got := meta["managed"]; got != true {
		t.Fatalf("meta[managed] = %v, want true", got)
	}
	if got := meta["relativePath"]; got != "demo/tab_1/20260317T060000Z-snapshot.json" {
		t.Fatalf("meta[relativePath] = %v, want managed relative path", got)
	}
	if got := meta["createdAt"]; got != "2026-03-17T06:00:00Z" {
		t.Fatalf("meta[createdAt] = %v, want RFC3339 timestamp", got)
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
	if got := meta["managed"]; got != true {
		t.Fatalf("meta[managed] = %v, want true", got)
	}
	if got := meta["relativePath"]; got != "demo/tab_1/20260317T060000Z-screenshot.jpg" {
		t.Fatalf("meta[relativePath] = %v, want managed relative path", got)
	}
	if got := meta["createdAt"]; got != "2026-03-17T06:00:00Z" {
		t.Fatalf("meta[createdAt] = %v, want RFC3339 timestamp", got)
	}
}

func TestSaveSnapshotArtifactExplicitPathIsMarkedUnmanaged(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	explicit := filepath.Join(t.TempDir(), "snapshot.json")
	meta, err := saveSnapshotArtifact(
		store,
		"demo",
		"tab_1",
		explicit,
		map[string]any{"nodes": []map[string]any{{"ref": "e1"}}},
		time.Date(2026, 3, 17, 6, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("saveSnapshotArtifact() error = %v", err)
	}
	if got := meta["managed"]; got != false {
		t.Fatalf("meta[managed] = %v, want false", got)
	}
	if _, exists := meta["relativePath"]; exists {
		t.Fatalf("meta contains relativePath for explicit artifact: %v", meta["relativePath"])
	}
}
