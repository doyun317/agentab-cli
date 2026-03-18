package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentab/agentab-cli/internal/state"
)

func artifactSessionName(store *state.Store, requested string) string {
	if requested != "" {
		return requested
	}
	current, err := store.CurrentSession()
	if err == nil && current.Name != "" {
		return current.Name
	}
	return "session-unspecified"
}

func sanitizeArtifactSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unspecified"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '.' || r == '_' || r == '-':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	sanitized := strings.Trim(builder.String(), "_")
	if sanitized == "" {
		return "unspecified"
	}
	return sanitized
}

func defaultArtifactPath(
	store *state.Store,
	sessionName string,
	tabID string,
	kind string,
	ext string,
	now time.Time,
) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	filename := fmt.Sprintf("%s-%s%s", now.UTC().Format("20060102T150405Z"), kind, ext)
	return filepath.Join(
		store.ArtifactsDir(),
		sanitizeArtifactSegment(sessionName),
		sanitizeArtifactSegment(tabID),
		filename,
	)
}

func writeArtifactFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func artifactMeta(
	store *state.Store,
	path string,
	data []byte,
	mimeType string,
	kind string,
	sessionName string,
	tabID string,
	now time.Time,
	managed bool,
) map[string]any {
	meta := map[string]any{
		"path":         path,
		"bytes":        len(data),
		"mimeType":     mimeType,
		"artifactKind": kind,
		"session":      sessionName,
		"tabId":        tabID,
		"managed":      managed,
		"createdAt":    now.UTC().Format(time.RFC3339),
	}
	if managed {
		if rel, err := filepath.Rel(store.ArtifactsDir(), path); err == nil {
			meta["relativePath"] = filepath.ToSlash(rel)
		}
	}
	return meta
}

func snapshotArtifactBytes(payload any) ([]byte, string, string, error) {
	if text, ok := payload.(string); ok {
		return []byte(text), "txt", "text/plain; charset=utf-8", nil
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, "", "", err
	}
	raw = append(raw, '\n')
	return raw, "json", "application/json", nil
}

func saveSnapshotArtifact(
	store *state.Store,
	sessionName string,
	tabID string,
	explicitPath string,
	payload any,
	now time.Time,
) (map[string]any, error) {
	data, ext, mimeType, err := snapshotArtifactBytes(payload)
	if err != nil {
		return nil, err
	}
	path := explicitPath
	managed := path == ""
	if path == "" {
		path = defaultArtifactPath(store, sessionName, tabID, "snapshot", ext, now)
	}
	if err := writeArtifactFile(path, data); err != nil {
		return nil, err
	}
	return artifactMeta(store, path, data, mimeType, "snapshot", sessionName, tabID, now, managed), nil
}

func saveBinaryArtifact(
	store *state.Store,
	sessionName string,
	tabID string,
	kind string,
	explicitPath string,
	defaultExt string,
	mimeType string,
	data []byte,
	now time.Time,
) (map[string]any, error) {
	path := explicitPath
	managed := path == ""
	if path == "" {
		path = defaultArtifactPath(store, sessionName, tabID, kind, defaultExt, now)
	}
	if err := writeArtifactFile(path, data); err != nil {
		return nil, err
	}
	return artifactMeta(store, path, data, mimeType, kind, sessionName, tabID, now, managed), nil
}
