package pinchtab

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentab/agentab-cli/internal/state"
)

func TestLaunchEnvWritesConfigWhenChromeBinPresent(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	chromePath := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(chromePath, []byte(""), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("CHROME_BIN", chromePath)

	env, err := launchEnv(store, "http://127.0.0.1:9867", "secret")
	if err != nil {
		t.Fatalf("launchEnv() error = %v", err)
	}

	foundConfig := ""
	for _, entry := range env {
		if strings.HasPrefix(entry, "PINCHTAB_CONFIG=") {
			foundConfig = strings.TrimPrefix(entry, "PINCHTAB_CONFIG=")
			break
		}
	}
	if foundConfig == "" {
		t.Fatal("PINCHTAB_CONFIG not found in environment")
	}

	raw, err := os.ReadFile(foundConfig)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", foundConfig, err)
	}
	if !strings.Contains(string(raw), chromePath) {
		t.Fatalf("config does not contain chrome path: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"port": "9867"`) {
		t.Fatalf("config does not contain port override: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"bind": "127.0.0.1"`) {
		t.Fatalf("config does not contain bind override: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"token": "secret"`) {
		t.Fatalf("config does not contain token override: %s", string(raw))
	}
}

func TestResolveRuntimeUsesStoredPinchtabInfo(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if err := store.WritePinchtabInfo(state.PinchtabInfo{
		BaseURL: "http://127.0.0.1:9988",
		Token:   "persisted-token",
	}); err != nil {
		t.Fatalf("WritePinchtabInfo() error = %v", err)
	}

	baseURL, token := resolveRuntime(store)
	if baseURL != "http://127.0.0.1:9988" {
		t.Fatalf("resolveRuntime() baseURL = %q, want persisted URL", baseURL)
	}
	if token != "persisted-token" {
		t.Fatalf("resolveRuntime() token = %q, want persisted-token", token)
	}
}

func TestResolveRuntimeFallsBackWhenDefaultPortIsOccupied(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:9867")
	if err != nil {
		t.Skip("default pinchtab port is already occupied on this machine")
	}
	defer ln.Close()

	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	baseURL, _ := resolveRuntime(store)
	if baseURL == "http://127.0.0.1:9867" {
		t.Fatalf("resolveRuntime() = %q, want fallback port when 9867 is occupied", baseURL)
	}
	if !strings.HasPrefix(baseURL, "http://127.0.0.1:") {
		t.Fatalf("resolveRuntime() = %q, want local URL", baseURL)
	}
}

func TestResolveRuntimePrefersEnvironmentURL(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	t.Setenv("PINCHTAB_URL", "http://127.0.0.1:9999")
	t.Setenv("PINCHTAB_TOKEN", "env-token")

	baseURL, token := resolveRuntime(store)
	if baseURL != "http://127.0.0.1:9999" {
		t.Fatalf("resolveRuntime() baseURL = %q, want env URL", baseURL)
	}
	if token != "env-token" {
		t.Fatalf("resolveRuntime() token = %q, want env-token", token)
	}
}

func TestResolveRuntimePersistsChosenLocalURL(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	baseURL, token := resolveRuntime(store)
	if token != "" {
		t.Fatalf("resolveRuntime() token = %q, want empty token", token)
	}

	info, err := store.ReadPinchtabInfo()
	if err != nil {
		t.Fatalf("ReadPinchtabInfo() error = %v", err)
	}
	if info.BaseURL != baseURL {
		t.Fatalf("persisted BaseURL = %q, want %q", info.BaseURL, baseURL)
	}
}
