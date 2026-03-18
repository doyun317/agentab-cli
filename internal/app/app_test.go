package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentab/agentab-cli/internal/response"
	"github.com/agentab/agentab-cli/internal/state"
)

func newStoreWithFakeDaemon(t *testing.T, handler http.HandlerFunc) *state.Store {
	t.Helper()

	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{"status": "running"}, nil))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	addr, ok := server.Listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("server.Listener.Addr() type = %T, want *net.TCPAddr", server.Listener.Addr())
	}
	if err := store.WriteDaemonInfo(state.DaemonInfo{
		Port:      addr.Port,
		Token:     "secret",
		PID:       1234,
		StartedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("WriteDaemonInfo() error = %v", err)
	}
	return store
}

func TestResolveChromeBinaryPrefersEnvOverride(t *testing.T) {
	chromePath := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(chromePath, []byte(""), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("CHROME_BIN", chromePath)

	info := resolveChromeBinary()
	if !info.Found {
		t.Fatalf("resolveChromeBinary().Found = false, want true")
	}
	if info.Source != "env" {
		t.Fatalf("resolveChromeBinary().Source = %q, want env", info.Source)
	}
	if info.Path != chromePath {
		t.Fatalf("resolveChromeBinary().Path = %q, want %q", info.Path, chromePath)
	}
	if info.Error != "" {
		t.Fatalf("resolveChromeBinary().Error = %q, want empty", info.Error)
	}
}

func TestResolveChromeBinaryReportsInvalidEnvOverride(t *testing.T) {
	t.Setenv("CHROME_BIN", "/missing/chrome")

	info := resolveChromeBinary()
	if info.Found {
		t.Fatalf("resolveChromeBinary().Found = true, want false")
	}
	if info.Source != "env" {
		t.Fatalf("resolveChromeBinary().Source = %q, want env", info.Source)
	}
	if info.Path != "/missing/chrome" {
		t.Fatalf("resolveChromeBinary().Path = %q, want /missing/chrome", info.Path)
	}
	if info.Error == "" {
		t.Fatal("resolveChromeBinary().Error = empty, want not found error")
	}
}

func TestRunDoctorReflectsChromeBinOverride(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	chromePath := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(chromePath, []byte(""), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	pinchtabPath := filepath.Join(t.TempDir(), "pinchtab")
	if err := os.WriteFile(pinchtabPath, []byte(""), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	t.Setenv("CHROME_BIN", chromePath)
	t.Setenv("AGENTAB_PINCHTAB_BIN", pinchtabPath)
	t.Setenv("PINCHTAB_URL", server.URL)

	env := runDoctor(context.Background(), store)
	if !env.OK {
		t.Fatalf("runDoctor() returned error: %+v", env.Error)
	}

	data, ok := env.Data.(doctorReport)
	if !ok {
		t.Fatalf("runDoctor().Data type = %T, want doctorReport", env.Data)
	}
	if got := data.ChromeBin; got != chromePath {
		t.Fatalf("runDoctor().Data.ChromeBin = %q, want %q", got, chromePath)
	}
	if got := data.ChromeBinSource; got != "env" {
		t.Fatalf("runDoctor().Data.ChromeBinSource = %q, want env", got)
	}
	if got := data.ChromeBinFound; got != true {
		t.Fatalf("runDoctor().Data.ChromeBinFound = %v, want true", got)
	}
	if data.ChromeBinError != "" {
		t.Fatalf("runDoctor().Data.ChromeBinError = %q, want empty", data.ChromeBinError)
	}
}

func TestDoctorReportRenderText(t *testing.T) {
	report := doctorReport{
		AgentabHome:     "/tmp/agentab",
		ArtifactsDir:    "/tmp/agentab/artifacts",
		ManagedBinPath:  "/tmp/agentab/bin/pinchtab",
		PinchtabURL:     "http://127.0.0.1:43921",
		PinchtabHealthy: true,
		PinchtabBin:     "/tmp/agentab/bin/pinchtab",
		PinchtabManaged: boolPtr(true),
		ChromeBin:       "/usr/bin/google-chrome",
		ChromeBinFound:  true,
		ChromeBinSource: "env",
		Daemon: &state.DaemonInfo{
			Port:      43921,
			Token:     "secret",
			PID:       1234,
			StartedAt: time.Date(2026, 3, 18, 3, 4, 5, 0, time.UTC),
		},
		Pinchtab: &state.PinchtabInfo{
			BaseURL:   "http://127.0.0.1:43921",
			Token:     "secret",
			PID:       5678,
			StartedAt: time.Date(2026, 3, 18, 3, 4, 7, 0, time.UTC),
		},
	}

	text := report.RenderText()
	for _, want := range []string{
		"agentab doctor",
		"home: /tmp/agentab",
		"chrome",
		"status: ok",
		"source: env",
		"path: /usr/bin/google-chrome",
		"pinchtab",
		"managed binary: yes",
		"runtime pid: 5678",
		"runtime token: present",
		"daemon",
		"status: running",
		"port: 43921",
		"token: present",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("RenderText() missing %q in:\n%s", want, text)
		}
	}
}

func TestRunSessionStopWithoutCurrentSessionReturnsNotFound(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	env := runSession(context.Background(), store, GlobalOptions{}, []string{"stop"})
	if env.OK {
		t.Fatal("runSession().OK = true, want false")
	}
	if env.Error == nil || env.Error.Code != "not_found" {
		t.Fatalf("runSession().Error = %#v, want not_found", env.Error)
	}
	if got := response.ExitCode(env); got != 4 {
		t.Fatalf("ExitCode() = %d, want 4", got)
	}
}

func TestRunSessionWithExplicitMissingSessionReturnsNotFound(t *testing.T) {
	store := newStoreWithFakeDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/sessions/missing/stop" {
			response.WriteJSON(w, http.StatusNotFound, response.Fail("not_found", "session missing not found", nil, nil))
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/sessions/missing/resume" {
			response.WriteJSON(w, http.StatusNotFound, response.Fail("not_found", "session missing not found", nil, nil))
			return
		}
		http.NotFound(w, r)
	})

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "stop", args: []string{"stop", "missing"}},
		{name: "resume", args: []string{"resume", "missing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := runSession(context.Background(), store, GlobalOptions{}, tc.args)
			if env.OK {
				t.Fatal("runSession().OK = true, want false")
			}
			if env.Error == nil || env.Error.Code != "not_found" {
				t.Fatalf("runSession().Error = %#v, want not_found", env.Error)
			}
			if env.Error.Message != "session missing not found" {
				t.Fatalf("runSession().Error.Message = %q, want %q", env.Error.Message, "session missing not found")
			}
			if got := response.ExitCode(env); got != 4 {
				t.Fatalf("ExitCode() = %d, want 4", got)
			}
		})
	}
}

func TestRunTabTextWithoutCurrentTabReturnsNotFound(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.PutSession(state.Session{
		Name:       "demo",
		InstanceID: "inst_1",
		LastUsedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("PutSession() error = %v", err)
	}

	env := runTab(context.Background(), store, GlobalOptions{Session: "demo"}, []string{"text"})
	if env.OK {
		t.Fatal("runTab().OK = true, want false")
	}
	if env.Error == nil || env.Error.Code != "not_found" {
		t.Fatalf("runTab().Error = %#v, want not_found", env.Error)
	}
	if got := response.ExitCode(env); got != 4 {
		t.Fatalf("ExitCode() = %d, want 4", got)
	}
}

func TestRunTabWithExplicitMissingIdentifiersReturnsNotFound(t *testing.T) {
	store := newStoreWithFakeDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/sessions/missing/tabs" {
			response.WriteJSON(w, http.StatusNotFound, response.Fail("not_found", "session missing not found", nil, nil))
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/tabs/tab_missing/text" {
			response.WriteJSON(w, http.StatusNotFound, response.Fail("not_found", "tab tab_missing not found", nil, nil))
			return
		}
		http.NotFound(w, r)
	})

	for _, tc := range []struct {
		name        string
		global      GlobalOptions
		args        []string
		wantMessage string
	}{
		{
			name:        "missing session on tab list",
			global:      GlobalOptions{Session: "missing"},
			args:        []string{"list"},
			wantMessage: "session missing not found",
		},
		{
			name:        "missing tab on tab text",
			global:      GlobalOptions{Tab: "tab_missing"},
			args:        []string{"text"},
			wantMessage: "tab tab_missing not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := runTab(context.Background(), store, tc.global, tc.args)
			if env.OK {
				t.Fatal("runTab().OK = true, want false")
			}
			if env.Error == nil || env.Error.Code != "not_found" {
				t.Fatalf("runTab().Error = %#v, want not_found", env.Error)
			}
			if env.Error.Message != tc.wantMessage {
				t.Fatalf("runTab().Error.Message = %q, want %q", env.Error.Message, tc.wantMessage)
			}
			if got := response.ExitCode(env); got != 4 {
				t.Fatalf("ExitCode() = %d, want 4", got)
			}
		})
	}
}

func TestRunTabClickReturnsLockConflictExitCode(t *testing.T) {
	store := newStoreWithFakeDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/tabs/tab_1/action" {
			response.WriteJSON(w, http.StatusConflict, response.Fail("lock_conflict", "tab tab_1 is locked by owner-1", nil, nil))
			return
		}
		http.NotFound(w, r)
	})
	if err := store.PutSession(state.Session{
		Name:         "demo",
		InstanceID:   "inst_1",
		CurrentTabID: "tab_1",
		LastUsedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("PutSession() error = %v", err)
	}

	env := runTab(context.Background(), store, GlobalOptions{Session: "demo"}, []string{"click", "--ref", "e1"})
	if env.OK {
		t.Fatal("runTab().OK = true, want false")
	}
	if env.Error == nil || env.Error.Code != "lock_conflict" {
		t.Fatalf("runTab().Error = %#v, want lock_conflict", env.Error)
	}
	if got := response.ExitCode(env); got != 5 {
		t.Fatalf("ExitCode() = %d, want 5", got)
	}
}

func TestRunSessionListReturnsTimeoutExitCode(t *testing.T) {
	store := newStoreWithFakeDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/sessions" {
			response.WriteJSON(w, http.StatusGatewayTimeout, response.Fail("timeout", "session list timed out", nil, nil))
			return
		}
		http.NotFound(w, r)
	})

	env := runSession(context.Background(), store, GlobalOptions{Timeout: time.Second}, []string{"list"})
	if env.OK {
		t.Fatal("runSession().OK = true, want false")
	}
	if env.Error == nil || env.Error.Code != "timeout" {
		t.Fatalf("runSession().Error = %#v, want timeout", env.Error)
	}
	if got := response.ExitCode(env); got != 6 {
		t.Fatalf("ExitCode() = %d, want 6", got)
	}
}

func TestRunSessionListReturnsUpstreamErrorExitCode(t *testing.T) {
	store := newStoreWithFakeDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/sessions" {
			response.WriteJSON(w, http.StatusBadGateway, response.Fail("upstream_error", "pinchtab upstream failed", nil, nil))
			return
		}
		http.NotFound(w, r)
	})

	env := runSession(context.Background(), store, GlobalOptions{Timeout: time.Second}, []string{"list"})
	if env.OK {
		t.Fatal("runSession().OK = true, want false")
	}
	if env.Error == nil || env.Error.Code != "upstream_error" {
		t.Fatalf("runSession().Error = %#v, want upstream_error", env.Error)
	}
	if got := response.ExitCode(env); got != 7 {
		t.Fatalf("ExitCode() = %d, want 7", got)
	}
}
