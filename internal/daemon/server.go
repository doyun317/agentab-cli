package daemon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/agentab/agentab-cli/internal/pinchtab"
	"github.com/agentab/agentab-cli/internal/response"
	"github.com/agentab/agentab-cli/internal/state"
)

type tabLock struct {
	Owner     string    `json:"owner"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Server struct {
	store   *state.Store
	manager *pinchtab.Manager
	token   string
	port    int

	lockMu sync.Mutex
	locks  map[string]tabLock

	srv *http.Server
}

func NewServer(store *state.Store, manager *pinchtab.Manager, token string, port int) *Server {
	return &Server{
		store:   store,
		manager: manager,
		token:   token,
		port:    port,
		locks:   map[string]tabLock{},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /shutdown", s.handleShutdown)
	mux.HandleFunc("GET /sessions", s.handleSessions)
	mux.HandleFunc("POST /sessions/start", s.handleSessionStart)
	mux.HandleFunc("GET /sessions/{name}", s.handleSessionGet)
	mux.HandleFunc("POST /sessions/{name}/resume", s.handleSessionResume)
	mux.HandleFunc("POST /sessions/{name}/stop", s.handleSessionStop)
	mux.HandleFunc("GET /sessions/{name}/tabs", s.handleTabsList)
	mux.HandleFunc("POST /sessions/{name}/tabs/open", s.handleTabOpen)
	mux.HandleFunc("POST /tabs/{id}/close", s.handleTabClose)
	mux.HandleFunc("POST /tabs/{id}/focus", s.handleTabFocus)
	mux.HandleFunc("GET /tabs/{id}/snapshot", s.handleTabSnapshot)
	mux.HandleFunc("GET /tabs/{id}/text", s.handleTabText)
	mux.HandleFunc("POST /tabs/{id}/find", s.handleTabFind)
	mux.HandleFunc("POST /tabs/{id}/action", s.handleTabAction)
	mux.HandleFunc("POST /tabs/{id}/evaluate", s.handleTabEvaluate)
	mux.HandleFunc("GET /tabs/{id}/screenshot", s.handleTabScreenshot)
	mux.HandleFunc("GET /tabs/{id}/pdf", s.handleTabPDF)
	return s.auth(mux)
}

func (s *Server) Serve(ctx context.Context) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
	if err != nil {
		return err
	}
	s.port = ln.Addr().(*net.TCPAddr).Port
	if err := s.store.WriteDaemonInfo(state.DaemonInfo{
		Port:      s.port,
		Token:     s.token,
		PID:       os.Getpid(),
		StartedAt: time.Now().UTC(),
	}); err != nil {
		_ = ln.Close()
		return err
	}

	s.srv = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Shutdown(shutdownCtx)
	}()

	err = s.srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	_ = s.manager.ShutdownOwned()
	_ = s.store.ClearDaemonInfo()
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token != "" {
			auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if auth != s.token {
				response.WriteJSON(w, http.StatusUnauthorized, response.Fail("dependency_error", "invalid daemon token", nil, nil))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	sessions, _ := s.store.ListSessions()
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"status":      "running",
		"port":        s.port,
		"sessions":    len(sessions),
		"pinchtabURL": s.manager.BaseURL(),
	}, nil))
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{"status": "shutting_down"}, nil))
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	}()
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.store.ListSessions()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, nil)
		return
	}
	st, _ := s.store.LoadState()
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"sessions":       sessions,
		"currentSession": st.CurrentSession,
	}, nil))
}

func (s *Server) handleSessionStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ProfileID string `json:"profileId"`
		Mode      string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "usage_error", err, nil)
		return
	}
	if req.Name == "" {
		req.Name = fmt.Sprintf("session-%d", time.Now().UTC().Unix())
	}
	if _, err := s.store.GetSession(req.Name); err == nil {
		s.writeError(w, http.StatusConflict, "usage_error", fmt.Errorf("session %s already exists", req.Name), nil)
		return
	}

	diag, err := s.manager.EnsureRunning(r.Context())
	if err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "dependency_error", err, nil)
		return
	}
	client := s.manager.Client(45 * time.Second)
	instance, err := client.StartInstance(r.Context(), pinchtab.StartInstanceRequest{
		ProfileID: req.ProfileID,
		Mode:      req.Mode,
	})
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, diag)
		return
	}

	session := state.Session{
		Name:       req.Name,
		InstanceID: instance.ID,
		ProfileID:  instance.ProfileID,
		Mode:       req.Mode,
		LastUsedAt: time.Now().UTC(),
	}
	if err := s.store.PutSession(session); err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, diag)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(session, diag))
}

func (s *Server) handleSessionGet(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.GetSession(r.PathValue("name"))
	if err != nil {
		s.writeError(w, http.StatusNotFound, "not_found", err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(session, nil))
}

func (s *Server) handleSessionResume(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	session, err := s.store.GetSession(name)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "not_found", err, nil)
		return
	}
	if err := s.store.SetCurrentSession(name); err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(session, nil))
}

func (s *Server) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	session, err := s.store.GetSession(name)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "not_found", err, nil)
		return
	}
	client := s.manager.Client(30 * time.Second)
	if err := client.StopInstance(r.Context(), session.InstanceID); err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	if err := s.store.DeleteSession(name); err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"name":       name,
		"instanceId": session.InstanceID,
		"stopped":    true,
	}, nil))
}

func (s *Server) handleTabsList(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.GetSession(r.PathValue("name"))
	if err != nil {
		s.writeError(w, http.StatusNotFound, "not_found", err, nil)
		return
	}
	client := s.manager.Client(30 * time.Second)
	tabs, err := listTabsWithCurrentRetry(r.Context(), client, session, 2*time.Second)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	out := make([]map[string]any, 0)
	for _, tab := range tabs {
		tabID := tab.ID
		if tabID == "" {
			tabID = tab.TabID
		}
		if tab.InstanceID != session.InstanceID {
			continue
		}
		lock := s.lockSnapshot(tabID)
		out = append(out, map[string]any{
			"tabId":         tabID,
			"session":       session.Name,
			"url":           tab.URL,
			"title":         tab.Title,
			"lockOwner":     lock.Owner,
			"lockExpiresAt": lock.ExpiresAt,
		})
	}
	slices.SortStableFunc(out, func(a, b map[string]any) int {
		aID, _ := a["tabId"].(string)
		bID, _ := b["tabId"].(string)
		if aID == session.CurrentTabID && bID != session.CurrentTabID {
			return -1
		}
		if bID == session.CurrentTabID && aID != session.CurrentTabID {
			return 1
		}
		return 0
	})
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"tabs":         out,
		"currentTabId": session.CurrentTabID,
	}, nil))
}

func (s *Server) handleTabOpen(w http.ResponseWriter, r *http.Request) {
	sessionName := r.PathValue("name")
	session, err := s.store.GetSession(sessionName)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "not_found", err, nil)
		return
	}
	var req struct {
		URL string `json:"url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.URL == "" {
		req.URL = "about:blank"
	}
	client := s.manager.Client(30 * time.Second)
	if err := waitForInstanceRunning(r.Context(), client, session.InstanceID, 12*time.Second); err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "dependency_error", err, nil)
		return
	}
	tab, err := client.OpenTab(r.Context(), session.InstanceID, req.URL)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	tabID := tab.ID
	if tabID == "" {
		tabID = tab.TabID
	}
	session, err = s.store.UpdateSession(sessionName, func(current *state.Session) error {
		current.CurrentTabID = tabID
		current.LastUsedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"tabId":   tabID,
		"session": session.Name,
		"url":     tab.URL,
		"title":   tab.Title,
	}, nil))
}

func waitForInstanceRunning(ctx context.Context, client *pinchtab.Client, instanceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		info, err := client.GetInstance(ctx, instanceID)
		if err == nil {
			switch info.Status {
			case "running":
				return nil
			case "error", "stopped":
				return fmt.Errorf("instance %s is %s", instanceID, info.Status)
			}
		}
		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("wait for instance %s: %w", instanceID, err)
			}
			return fmt.Errorf("instance %s did not become ready in %s", instanceID, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func listTabsWithCurrentRetry(ctx context.Context, client *pinchtab.Client, session state.Session, timeout time.Duration) ([]pinchtab.TabInfo, error) {
	deadline := time.Now().Add(timeout)
	for {
		tabs, err := client.ListTabs(ctx, session.InstanceID)
		if err != nil {
			return nil, err
		}
		if session.CurrentTabID == "" || containsTabID(tabs, session.CurrentTabID) || time.Now().After(deadline) {
			return tabs, nil
		}
		select {
		case <-ctx.Done():
			return tabs, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func containsTabID(tabs []pinchtab.TabInfo, tabID string) bool {
	for _, tab := range tabs {
		id := tab.ID
		if id == "" {
			id = tab.TabID
		}
		if id == tabID {
			return true
		}
	}
	return false
}

func (s *Server) handleTabClose(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("id")
	if err := s.manager.Client(30*time.Second).CloseTab(r.Context(), tabID); err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	s.clearTabFromSessions(tabID)
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{"tabId": tabID, "closed": true}, nil))
}

func (s *Server) handleTabFocus(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("id")
	sessionName := r.URL.Query().Get("session")
	if sessionName == "" {
		s.writeError(w, http.StatusBadRequest, "usage_error", fmt.Errorf("session query is required"), nil)
		return
	}
	if err := s.manager.Client(30*time.Second).FocusTab(r.Context(), tabID); err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	session, err := s.store.UpdateSession(sessionName, func(current *state.Session) error {
		current.CurrentTabID = tabID
		current.LastUsedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "upstream_error", err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(session, nil))
}

func (s *Server) handleTabSnapshot(w http.ResponseWriter, r *http.Request) {
	result, err := s.manager.Client(30*time.Second).Snapshot(r.Context(), r.PathValue("id"), pinchtab.SnapshotParams{
		Filter:    r.URL.Query().Get("filter"),
		Format:    r.URL.Query().Get("format"),
		Selector:  r.URL.Query().Get("selector"),
		MaxTokens: r.URL.Query().Get("maxTokens"),
		Depth:     r.URL.Query().Get("depth"),
		Diff:      r.URL.Query().Get("diff") == "true",
	})
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(result, nil))
}

func (s *Server) handleTabText(w http.ResponseWriter, r *http.Request) {
	result, err := s.manager.Client(30*time.Second).Text(r.Context(), r.PathValue("id"), r.URL.Query().Get("mode"))
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(result, nil))
}

func (s *Server) handleTabFind(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query     string `json:"query"`
		Threshold string `json:"threshold"`
		Explain   bool   `json:"explain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "usage_error", err, nil)
		return
	}
	result, err := s.manager.Client(30*time.Second).Find(r.Context(), r.PathValue("id"), req.Query, req.Threshold, req.Explain)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(result, nil))
}

func (s *Server) handleTabAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Owner    string `json:"owner"`
		Kind     string `json:"kind"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
		Text     string `json:"text"`
		Key      string `json:"key"`
		Value    string `json:"value"`
		ScrollY  int    `json:"scrollY"`
		WaitNav  bool   `json:"waitNav"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "usage_error", err, nil)
		return
	}
	tabID := r.PathValue("id")
	owner := req.Owner
	if owner == "" {
		owner = "agentab"
	}
	if err := s.acquireLock(tabID, owner, 30*time.Second); err != nil {
		s.writeError(w, http.StatusConflict, "lock_conflict", err, nil)
		return
	}
	defer s.releaseLock(tabID, owner)

	result, err := s.manager.Client(30*time.Second).Action(r.Context(), tabID, pinchtab.ActionRequest{
		Kind:     req.Kind,
		Ref:      req.Ref,
		Selector: req.Selector,
		Text:     req.Text,
		Key:      req.Key,
		Value:    req.Value,
		ScrollY:  req.ScrollY,
		WaitNav:  req.WaitNav,
	})
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(result, nil))
}

func (s *Server) handleTabEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "usage_error", err, nil)
		return
	}
	result, err := s.manager.Client(30*time.Second).Evaluate(r.Context(), r.PathValue("id"), req.Expression)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(result, nil))
}

func (s *Server) handleTabScreenshot(w http.ResponseWriter, r *http.Request) {
	data, err := s.manager.Client(30*time.Second).Screenshot(r.Context(), r.PathValue("id"), r.URL.Query().Get("quality"))
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"mimeType": "image/jpeg",
		"data":     base64.StdEncoding.EncodeToString(data),
		"bytes":    len(data),
	}, nil))
}

func (s *Server) handleTabPDF(w http.ResponseWriter, r *http.Request) {
	data, err := s.manager.Client(30*time.Second).PDF(r.Context(), r.PathValue("id"), r.URL.Query().Get("scale"), r.URL.Query().Get("landscape") == "true")
	if err != nil {
		s.writeError(w, http.StatusBadGateway, s.errorCode(err), err, nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.OK(map[string]any{
		"mimeType": "application/pdf",
		"data":     base64.StdEncoding.EncodeToString(data),
		"bytes":    len(data),
	}, nil))
}

func (s *Server) acquireLock(tabID, owner string, ttl time.Duration) error {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if current, ok := s.locks[tabID]; ok && time.Now().Before(current.ExpiresAt) && current.Owner != owner {
		return fmt.Errorf("tab %s is locked by %s", tabID, current.Owner)
	}
	s.locks[tabID] = tabLock{Owner: owner, ExpiresAt: time.Now().Add(ttl)}
	return nil
}

func (s *Server) releaseLock(tabID, owner string) {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	current, ok := s.locks[tabID]
	if !ok || current.Owner != owner {
		return
	}
	delete(s.locks, tabID)
}

func (s *Server) lockSnapshot(tabID string) tabLock {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	current, ok := s.locks[tabID]
	if !ok {
		return tabLock{}
	}
	if time.Now().After(current.ExpiresAt) {
		delete(s.locks, tabID)
		return tabLock{}
	}
	return current
}

func (s *Server) clearTabFromSessions(tabID string) {
	sessions, err := s.store.ListSessions()
	if err != nil {
		return
	}
	for _, session := range sessions {
		if session.CurrentTabID != tabID {
			continue
		}
		_, _ = s.store.UpdateSession(session.Name, func(current *state.Session) error {
			current.CurrentTabID = ""
			current.LastUsedAt = time.Now().UTC()
			return nil
		})
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, code string, err error, diagnostics map[string]any) {
	message := err.Error()
	if errors.Is(err, state.ErrNotFound) {
		code = "not_found"
	}
	response.WriteJSON(w, status, response.Fail(code, message, nil, diagnostics))
}

func (s *Server) errorCode(err error) string {
	var apiErr *pinchtab.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusNotFound:
			return "not_found"
		case http.StatusRequestTimeout, http.StatusGatewayTimeout:
			return "timeout"
		default:
			return "upstream_error"
		}
	}
	return "upstream_error"
}
