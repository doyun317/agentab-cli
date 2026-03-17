package pinchtab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pinchtab api error: status=%d body=%s", e.Status, e.Body)
}

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type StartInstanceRequest struct {
	ProfileID string `json:"profileId,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

type InstanceInfo struct {
	ID          string `json:"id"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	Port        string `json:"port"`
	Headless    bool   `json:"headless"`
	Status      string `json:"status"`
}

type TabInfo struct {
	ID         string `json:"id,omitempty"`
	TabID      string `json:"tabId,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	URL        string `json:"url,omitempty"`
	Title      string `json:"title,omitempty"`
}

type SnapshotParams struct {
	Filter    string
	Format    string
	Selector  string
	MaxTokens string
	Depth     string
	Diff      bool
}

type ActionRequest struct {
	Kind     string `json:"kind"`
	Ref      string `json:"ref,omitempty"`
	Selector string `json:"selector,omitempty"`
	Text     string `json:"text,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	ScrollY  int    `json:"scrollY,omitempty"`
	WaitNav  bool   `json:"waitNav,omitempty"`
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, http.MethodGet, "/health", nil, nil, &out)
	return out, err
}

func (c *Client) StartInstance(ctx context.Context, req StartInstanceRequest) (InstanceInfo, error) {
	var out InstanceInfo
	err := c.doJSON(ctx, http.MethodPost, "/instances/start", nil, req, &out)
	return out, err
}

func (c *Client) StopInstance(ctx context.Context, id string) error {
	var out map[string]any
	return c.doJSON(ctx, http.MethodPost, "/instances/"+id+"/stop", nil, nil, &out)
}

func (c *Client) Shutdown(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/shutdown", nil)
	if err != nil {
		return err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return &APIError{Status: resp.StatusCode, Body: string(body)}
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	healthCtx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	if _, healthErr := c.Health(healthCtx); healthErr != nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func (c *Client) GetInstance(ctx context.Context, id string) (InstanceInfo, error) {
	var out InstanceInfo
	err := c.doJSON(ctx, http.MethodGet, "/instances/"+id, nil, nil, &out)
	return out, err
}

func (c *Client) ListTabs(ctx context.Context, instanceID string) ([]TabInfo, error) {
	path := "/tabs"
	if instanceID != "" {
		path = "/instances/" + instanceID + "/tabs"
	}
	raw, err := c.do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}

	var wrapped struct {
		Tabs []TabInfo `json:"tabs"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Tabs != nil {
		return wrapped.Tabs, nil
	}

	var tabs []TabInfo
	if err := json.Unmarshal(raw, &tabs); err != nil {
		return nil, err
	}
	return tabs, nil
}

func (c *Client) OpenTab(ctx context.Context, instanceID, rawURL string) (TabInfo, error) {
	body := map[string]any{"url": rawURL}
	var out TabInfo
	err := c.doJSON(ctx, http.MethodPost, "/instances/"+instanceID+"/tabs/open", nil, body, &out)
	if out.ID == "" {
		out.ID = out.TabID
	}
	return out, err
}

func (c *Client) CloseTab(ctx context.Context, tabID string) error {
	var out map[string]any
	return c.doJSON(ctx, http.MethodPost, "/tabs/"+tabID+"/close", nil, nil, &out)
}

func (c *Client) FocusTab(ctx context.Context, tabID string) error {
	body := map[string]any{"action": "focus", "tabId": tabID}
	var out map[string]any
	return c.doJSON(ctx, http.MethodPost, "/tab", nil, body, &out)
}

func (c *Client) Snapshot(ctx context.Context, tabID string, params SnapshotParams) (any, error) {
	query := url.Values{}
	if params.Filter != "" {
		query.Set("filter", params.Filter)
	}
	if params.Format != "" {
		query.Set("format", params.Format)
	}
	if params.Selector != "" {
		query.Set("selector", params.Selector)
	}
	if params.MaxTokens != "" {
		query.Set("maxTokens", params.MaxTokens)
	}
	if params.Depth != "" {
		query.Set("depth", params.Depth)
	}
	if params.Diff {
		query.Set("diff", "true")
	}
	return c.getAny(ctx, "/tabs/"+tabID+"/snapshot", query)
}

func (c *Client) Text(ctx context.Context, tabID, mode string) (any, error) {
	query := url.Values{}
	if mode != "" {
		query.Set("mode", mode)
	}
	return c.getAny(ctx, "/tabs/"+tabID+"/text", query)
}

func (c *Client) Find(ctx context.Context, tabID, queryText string, threshold string, explain bool) (map[string]any, error) {
	body := map[string]any{"query": queryText}
	if threshold != "" {
		body["threshold"] = threshold
	}
	if explain {
		body["explain"] = true
	}
	var out map[string]any
	err := c.doJSON(ctx, http.MethodPost, "/tabs/"+tabID+"/find", nil, body, &out)
	return out, err
}

func (c *Client) Action(ctx context.Context, tabID string, req ActionRequest) (any, error) {
	return c.postAny(ctx, "/tabs/"+tabID+"/action", req)
}

func (c *Client) Evaluate(ctx context.Context, tabID, expression string) (any, error) {
	return c.postAny(ctx, "/tabs/"+tabID+"/evaluate", map[string]any{"expression": expression})
}

func (c *Client) Screenshot(ctx context.Context, tabID, quality string) ([]byte, error) {
	query := url.Values{}
	query.Set("raw", "true")
	if quality != "" {
		query.Set("quality", quality)
	}
	return c.doBytes(ctx, http.MethodGet, "/tabs/"+tabID+"/screenshot", query, nil)
}

func (c *Client) PDF(ctx context.Context, tabID, scale string, landscape bool) ([]byte, error) {
	query := url.Values{}
	query.Set("raw", "true")
	if scale != "" {
		query.Set("scale", scale)
	}
	if landscape {
		query.Set("landscape", "true")
	}
	return c.doBytes(ctx, http.MethodGet, "/tabs/"+tabID+"/pdf", query, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	raw, err := c.do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) getAny(ctx context.Context, path string, query url.Values) (any, error) {
	raw, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	return decodeAny(raw)
}

func (c *Client) postAny(ctx context.Context, path string, body any) (any, error) {
	raw, err := c.do(ctx, http.MethodPost, path, nil, body)
	if err != nil {
		return nil, err
	}
	return decodeAny(raw)
}

func (c *Client) doBytes(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	return c.do(ctx, method, path, query, body)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	target := c.BaseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, payload)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &APIError{Status: resp.StatusCode, Body: string(raw)}
	}
	return raw, nil
}

func decodeAny(raw []byte) (any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return string(raw), nil
	}
	return out, nil
}
