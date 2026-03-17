package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentab/agentab-cli/internal/pinchtab"
	"github.com/agentab/agentab-cli/internal/state"
)

type fakePinchTab struct {
	instanceID     string
	tabID          string
	lastAction     map[string]any
	instanceChecks int
	tabListCalls   int
	staleFirstList bool
}

func (f *fakePinchTab) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON := func(value any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(value)
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		writeJSON(map[string]any{"status": "ok"})
	case r.Method == http.MethodPost && r.URL.Path == "/instances/start":
		f.instanceID = "inst_1"
		writeJSON(map[string]any{
			"id":          f.instanceID,
			"profileId":   "prof_1",
			"profileName": "default",
			"port":        "9868",
			"headless":    true,
			"status":      "running",
		})
	case r.Method == http.MethodGet && r.URL.Path == "/instances/inst_1":
		f.instanceChecks++
		status := "running"
		if f.instanceChecks == 1 {
			status = "starting"
		}
		writeJSON(map[string]any{
			"id":          f.instanceID,
			"profileId":   "prof_1",
			"profileName": "default",
			"port":        "9868",
			"headless":    true,
			"status":      status,
		})
	case r.Method == http.MethodPost && r.URL.Path == "/instances/inst_1/tabs/open":
		f.tabID = "tab_1"
		writeJSON(map[string]any{
			"tabId": f.tabID,
			"url":   "https://example.com",
			"title": "Example",
		})
	case r.Method == http.MethodGet && r.URL.Path == "/instances/inst_1/tabs":
		f.tabListCalls++
		if f.staleFirstList && f.tabListCalls == 1 {
			writeJSON([]map[string]any{{
				"id":         "tab_blank",
				"instanceId": f.instanceID,
				"url":        "about:blank",
				"title":      "about:blank",
			}})
			return
		}
		writeJSON([]map[string]any{
			{
				"id":         "tab_blank",
				"instanceId": f.instanceID,
				"url":        "about:blank",
				"title":      "about:blank",
			},
			{
				"id":         f.tabID,
				"instanceId": f.instanceID,
				"url":        "https://example.com",
				"title":      "Example",
			},
		})
	case r.Method == http.MethodGet && r.URL.Path == "/tabs/tab_1/snapshot":
		writeJSON(map[string]any{"nodes": []map[string]any{{"ref": "e1", "role": "link", "name": "Example"}}})
	case r.Method == http.MethodPost && r.URL.Path == "/tabs/tab_1/find":
		writeJSON(map[string]any{"best_ref": "e1", "matches": []map[string]any{{"ref": "e1", "score": 0.99}}})
	case r.Method == http.MethodPost && r.URL.Path == "/tabs/tab_1/action":
		_ = json.NewDecoder(r.Body).Decode(&f.lastAction)
		writeJSON(map[string]any{"status": "ok"})
	case r.Method == http.MethodGet && r.URL.Path == "/tabs/tab_1/text":
		writeJSON(map[string]any{"text": "hello world"})
	case r.Method == http.MethodGet && r.URL.Path == "/tabs/tab_1/screenshot":
		_, _ = w.Write([]byte("jpeg-bytes"))
	case r.Method == http.MethodGet && r.URL.Path == "/tabs/tab_1/pdf":
		_, _ = w.Write([]byte("pdf-bytes"))
	case r.Method == http.MethodPost && r.URL.Path == "/tabs/tab_1/close":
		writeJSON(map[string]any{"closed": true})
	case r.Method == http.MethodPost && r.URL.Path == "/instances/inst_1/stop":
		writeJSON(map[string]any{"stopped": true})
	case r.Method == http.MethodPost && r.URL.Path == "/tab":
		writeJSON(map[string]any{"focused": true})
	default:
		http.NotFound(w, r)
	}
}

func TestServerSessionAndTabFlow(t *testing.T) {
	fake := &fakePinchTab{}
	upstream := httptest.NewServer(fake)
	defer upstream.Close()

	t.Setenv("PINCHTAB_URL", upstream.URL)
	t.Setenv("AGENTAB_SKIP_INSTALL", "1")

	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	server := NewServer(store, pinchtab.NewManager(store), "secret", 0)
	api := httptest.NewServer(server.Handler())
	defer api.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	call := func(method, path string, body any) map[string]any {
		var payload []byte
		if body != nil {
			payload, _ = json.Marshal(body)
		}
		req, _ := http.NewRequest(method, api.URL+path, bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer secret")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %s %s error = %v", method, path, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			t.Fatalf("request %s %s status = %d", method, path, resp.StatusCode)
		}
		var env map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatalf("decode %s %s error = %v", method, path, err)
		}
		return env
	}

	call(http.MethodPost, "/sessions/start", map[string]any{"name": "demo"})
	call(http.MethodPost, "/sessions/demo/tabs/open", map[string]any{"url": "https://example.com"})
	tabsEnv := call(http.MethodGet, "/sessions/demo/tabs", nil)
	call(http.MethodGet, "/tabs/tab_1/snapshot?filter=interactive&format=compact", nil)
	call(http.MethodPost, "/tabs/tab_1/find", map[string]any{"query": "Example"})
	call(http.MethodPost, "/tabs/tab_1/action", map[string]any{"owner": "test-agent", "kind": "click", "ref": "e1"})
	call(http.MethodGet, "/tabs/tab_1/text", nil)
	call(http.MethodGet, "/tabs/tab_1/screenshot", nil)
	call(http.MethodGet, "/tabs/tab_1/pdf", nil)
	call(http.MethodPost, "/tabs/tab_1/close", nil)
	call(http.MethodPost, "/sessions/demo/stop", nil)

	data, ok := tabsEnv["data"].(map[string]any)
	if !ok {
		t.Fatalf("tabs data = %#v, want object", tabsEnv["data"])
	}
	tabs, ok := data["tabs"].([]any)
	if !ok || len(tabs) != 2 {
		t.Fatalf("tabs = %#v, want two tabs", data["tabs"])
	}
	tab, ok := tabs[0].(map[string]any)
	if !ok || tab["tabId"] != "tab_1" {
		t.Fatalf("first tab entry = %#v, want tab_1 as current tab", tabs[0])
	}
	if data["currentTabId"] != "tab_1" {
		t.Fatalf("currentTabId = %#v, want tab_1", data["currentTabId"])
	}

	if fake.lastAction["kind"] != "click" || fake.lastAction["ref"] != "e1" {
		t.Fatalf("lastAction = %+v, want click/e1", fake.lastAction)
	}
	if fake.instanceChecks == 0 {
		t.Fatal("expected instance readiness checks before opening a tab")
	}
	if sessions, err := store.ListSessions(); err != nil || len(sessions) != 0 {
		t.Fatalf("store.ListSessions() = %+v, %v; want empty", sessions, err)
	}
}

func TestServerTabsListRetriesUntilCurrentTabAppears(t *testing.T) {
	fake := &fakePinchTab{staleFirstList: true}
	upstream := httptest.NewServer(fake)
	defer upstream.Close()

	t.Setenv("PINCHTAB_URL", upstream.URL)
	t.Setenv("AGENTAB_SKIP_INSTALL", "1")

	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	server := NewServer(store, pinchtab.NewManager(store), "secret", 0)
	api := httptest.NewServer(server.Handler())
	defer api.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	call := func(method, path string, body any) map[string]any {
		var payload []byte
		if body != nil {
			payload, _ = json.Marshal(body)
		}
		req, _ := http.NewRequest(method, api.URL+path, bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer secret")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %s %s error = %v", method, path, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			t.Fatalf("request %s %s status = %d", method, path, resp.StatusCode)
		}
		var env map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatalf("decode %s %s error = %v", method, path, err)
		}
		return env
	}

	call(http.MethodPost, "/sessions/start", map[string]any{"name": "demo"})
	call(http.MethodPost, "/sessions/demo/tabs/open", map[string]any{"url": "https://example.com"})
	tabsEnv := call(http.MethodGet, "/sessions/demo/tabs", nil)

	data, ok := tabsEnv["data"].(map[string]any)
	if !ok {
		t.Fatalf("tabs data = %#v, want object", tabsEnv["data"])
	}
	tabs, ok := data["tabs"].([]any)
	if !ok || len(tabs) < 2 {
		t.Fatalf("tabs = %#v, want current tab after retry", data["tabs"])
	}
	tab, ok := tabs[0].(map[string]any)
	if !ok || tab["tabId"] != "tab_1" {
		t.Fatalf("first tab entry = %#v, want tab_1 after retry", tabs[0])
	}
	if fake.tabListCalls < 2 {
		t.Fatalf("tabListCalls = %d, want retry", fake.tabListCalls)
	}
}

func TestServerLockConflict(t *testing.T) {
	t.Setenv("PINCHTAB_URL", "http://127.0.0.1:1")

	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	server := NewServer(store, pinchtab.NewManager(store), "secret", 0)
	if err := server.acquireLock("tab_1", "owner-1", time.Minute); err != nil {
		t.Fatalf("acquireLock() error = %v", err)
	}

	reqBody := bytes.NewBufferString(`{"owner":"owner-2","kind":"click","ref":"e1"}`)
	req := httptest.NewRequest(http.MethodPost, "/tabs/tab_1/action", reqBody)
	req.SetPathValue("id", "tab_1")
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestServerServeShutdownClearsDaemonInfo(t *testing.T) {
	store, err := state.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	server := NewServer(store, pinchtab.NewManager(store), "secret", 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Serve(ctx)
	}()

	time.Sleep(250 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve() did not stop")
	}

	if _, err := store.ReadDaemonInfo(); err != state.ErrNotFound {
		t.Fatalf("ReadDaemonInfo() after shutdown error = %v, want %v", err, state.ErrNotFound)
	}
}
