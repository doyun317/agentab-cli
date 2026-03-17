package daemon

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/agentab/agentab-cli/internal/response"
	"github.com/agentab/agentab-cli/internal/state"
)

const DefaultPort = 43921

type Client struct {
	store      *state.Store
	timeout    time.Duration
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(store *state.Store, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		store:      store,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Ensure(ctx context.Context) error {
	info, err := c.store.ReadDaemonInfo()
	if err == nil {
		if c.ping(ctx, info.Port, info.Token) == nil {
			c.setRuntime(info.Port, info.Token)
			return nil
		}
	}

	token := ""
	port := DefaultPort
	if err == nil {
		token = info.Token
		if info.Port > 0 {
			port = info.Port
		}
	}
	if token == "" {
		token, err = randomToken()
		if err != nil {
			return err
		}
	}
	port, err = selectDaemonPort(port)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	logPath := filepath.Join(c.store.LogsDir(), "agentab-daemon.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "daemon", "serve", "--port", strconv.Itoa(port), "--token", token)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = os.Environ()
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	go func() {
		_ = cmd.Process.Release()
		_ = logFile.Close()
	}()

	info = state.DaemonInfo{
		Port:      port,
		Token:     token,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now().UTC(),
	}
	if err := c.store.WriteDaemonInfo(info); err != nil {
		return err
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if c.ping(ctx, port, token) == nil {
			c.setRuntime(port, token)
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("agentab daemon did not start; see %s", logPath)
}

func (c *Client) Request(ctx context.Context, method, path string, body any) (response.Envelope, error) {
	if err := c.Ensure(ctx); err != nil {
		return response.Fail("dependency_error", err.Error(), nil, nil), err
	}
	return c.request(ctx, method, path, body)
}

func (c *Client) Status(ctx context.Context) (response.Envelope, error) {
	info, err := c.store.ReadDaemonInfo()
	if err != nil {
		return response.Fail("not_found", "daemon metadata not found", nil, nil), err
	}
	c.setRuntime(info.Port, info.Token)
	return c.request(ctx, http.MethodGet, "/health", nil)
}

func (c *Client) Stop(ctx context.Context) (response.Envelope, error) {
	info, err := c.store.ReadDaemonInfo()
	if err != nil {
		return response.Fail("not_found", "daemon metadata not found", nil, nil), err
	}
	c.setRuntime(info.Port, info.Token)
	env, reqErr := c.request(ctx, http.MethodPost, "/shutdown", nil)
	_ = c.store.ClearDaemonInfo()
	return env, reqErr
}

func (c *Client) setRuntime(port int, token string) {
	c.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	c.token = token
}

func (c *Client) ping(ctx context.Context, port int, token string) error {
	c.setRuntime(port, token)
	_, err := c.request(ctx, http.MethodGet, "/health", nil)
	return err
}

func (c *Client) request(ctx context.Context, method, path string, body any) (response.Envelope, error) {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil), err
		}
		payload = bytesReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return response.Fail("usage_error", err.Error(), nil, nil), err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response.Fail("dependency_error", err.Error(), nil, nil), err
	}
	defer resp.Body.Close()

	var env response.Envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return response.Fail("upstream_error", "failed to decode daemon response", nil, nil), err
	}
	if resp.StatusCode >= 400 {
		return env, fmt.Errorf("daemon status %d", resp.StatusCode)
	}
	return env, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func bytesReader(raw []byte) io.Reader {
	return bytes.NewReader(raw)
}

func selectDaemonPort(preferred int) (int, error) {
	if preferred > 0 {
		if port, ok := reservePort(preferred); ok {
			return port, nil
		}
	}
	port, ok := reservePort(0)
	if !ok {
		return 0, fmt.Errorf("allocate daemon port")
	}
	return port, nil
}

func reservePort(port int) (int, bool) {
	addr := "127.0.0.1:0"
	if port > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, false
	}
	defer ln.Close()

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok || tcpAddr.Port <= 0 {
		return 0, false
	}
	return tcpAddr.Port, true
}
