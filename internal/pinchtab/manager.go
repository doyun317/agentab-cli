package pinchtab

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/agentab/agentab-cli/internal/install"
	"github.com/agentab/agentab-cli/internal/state"
)

type Manager struct {
	store     *state.Store
	installer *install.Installer
	baseURL   string
	token     string

	mu       sync.Mutex
	ownedCmd *exec.Cmd
}

func NewManager(store *state.Store) *Manager {
	baseURL, token := resolveRuntime(store)
	return &Manager{
		store:     store,
		installer: install.New(store.Root()),
		baseURL:   strings.TrimRight(baseURL, "/"),
		token:     token,
	}
}

func (m *Manager) BaseURL() string { return m.baseURL }
func (m *Manager) Token() string   { return m.token }

func (m *Manager) Client(timeout time.Duration) *Client {
	return NewClient(m.baseURL, m.token, timeout)
}

func (m *Manager) EnsureRunning(ctx context.Context) (map[string]any, error) {
	client := m.Client(1500 * time.Millisecond)
	if _, err := client.Health(ctx); err == nil {
		if os.Getenv("PINCHTAB_URL") == "" {
			_ = m.store.WritePinchtabInfo(state.PinchtabInfo{
				BaseURL: m.baseURL,
				Token:   m.token,
			})
		}
		return map[string]any{
			"pinchtabURL": m.baseURL,
			"pinchtab":    "reachable",
			"source":      "existing",
		}, nil
	}

	if os.Getenv("PINCHTAB_URL") != "" && !isLocalBaseURL(m.baseURL) {
		return nil, fmt.Errorf("remote pinchtab is unreachable at %s", m.baseURL)
	}

	result, err := m.installer.EnsureInstalled(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.startProcess(result.Path); err != nil {
		return nil, err
	}
	return map[string]any{
		"pinchtabURL": m.baseURL,
		"pinchtab":    "launched",
		"binary":      result.Path,
		"version":     result.Version,
		"source":      "managed",
	}, nil
}

func (m *Manager) ShutdownOwned() error {
	m.mu.Lock()
	ownedCmd := m.ownedCmd
	m.ownedCmd = nil
	m.mu.Unlock()

	if ownedCmd == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Client(2 * time.Second).Shutdown(ctx); err == nil {
		return nil
	}
	if ownedCmd.Process == nil {
		return nil
	}
	return ownedCmd.Process.Kill()
}

func (m *Manager) startProcess(binary string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logPath := filepath.Join(m.store.LogsDir(), "pinchtab.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open pinchtab log: %w", err)
	}

	cmd := exec.Command(binary, "server")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	env, err := launchEnv(m.store, m.baseURL, m.token)
	if err != nil {
		_ = logFile.Close()
		return err
	}
	cmd.Env = append(os.Environ(), env...)
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("start pinchtab: %w", err)
	}

	m.ownedCmd = cmd
	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()
	_ = m.store.WritePinchtabInfo(state.PinchtabInfo{
		BaseURL:   m.baseURL,
		Token:     m.token,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now().UTC(),
	})

	healthClient := m.Client(1500 * time.Millisecond)
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := healthClient.Health(context.Background()); err == nil {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	_ = m.store.ClearPinchtabInfo()
	return fmt.Errorf("pinchtab did not become healthy; see %s", logPath)
}

func launchEnv(store *state.Store, baseURL, token string) ([]string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	host := u.Hostname()
	port := u.Port()
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "9867"
	}

	env := []string{
		"PINCHTAB_BIND=" + host,
		"PINCHTAB_PORT=" + port,
	}
	if token != "" {
		env = append(env, "PINCHTAB_TOKEN="+token)
	}

	if chromeBinary := os.Getenv("CHROME_BIN"); chromeBinary != "" {
		configPath := filepath.Join(store.RunDir(), "pinchtab-config.json")
		if err := writePinchtabConfig(configPath, host, port, token, chromeBinary); err != nil {
			return nil, err
		}
		env = append(env, "PINCHTAB_CONFIG="+configPath)
	}

	return env, nil
}

func isLocalBaseURL(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "127.0.0.1" || host == "localhost" || host == ""
}

func writePinchtabConfig(path, host, port, token, chromeBinary string) error {
	server := map[string]any{
		"bind": host,
		"port": port,
	}
	if token != "" {
		server["token"] = token
	}

	payload := map[string]any{
		"server": server,
		"browser": map[string]any{
			"binary": chromeBinary,
		},
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o600)
}

func resolveRuntime(store *state.Store) (string, string) {
	if baseURL := strings.TrimRight(os.Getenv("PINCHTAB_URL"), "/"); baseURL != "" {
		return baseURL, os.Getenv("PINCHTAB_TOKEN")
	}

	if info, err := store.ReadPinchtabInfo(); err == nil && info.BaseURL != "" {
		token := os.Getenv("PINCHTAB_TOKEN")
		if token == "" {
			token = info.Token
		}
		return strings.TrimRight(info.BaseURL, "/"), token
	}

	port, err := reservePort(0)
	if err != nil {
		port = 9867
		if selected, reserveErr := reservePort(port); reserveErr == nil {
			port = selected
		}
	}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	token := os.Getenv("PINCHTAB_TOKEN")
	_ = store.WritePinchtabInfo(state.PinchtabInfo{
		BaseURL: baseURL,
		Token:   token,
	})
	return baseURL, token
}

func reservePort(port int) (int, error) {
	addr := "127.0.0.1:0"
	if port > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok || tcpAddr.Port <= 0 {
		return 0, fmt.Errorf("resolve pinchtab port")
	}
	return tcpAddr.Port, nil
}
