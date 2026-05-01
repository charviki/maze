package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL   string
	authToken string
	client    *http.Client
}

func NewAPIClient(managerURL, authToken string) *APIClient {
	return &APIClient{
		baseURL:   managerURL,
		authToken: authToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateHost 创建 Host，断言返回 202，打印耗时
func (c *APIClient) CreateHost(name string, tools []string) (*HostInfo, error) {
	start := time.Now()
	reqBody := CreateHostRequest{
		Name:  name,
		Tools: tools,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", "/api/v1/hosts", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create host: expected 202, got %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var info HostInfo
	if err := json.Unmarshal(dataBytes, &info); err != nil {
		return nil, fmt.Errorf("unmarshal host info: %w", err)
	}
	log.Printf("  [api] CreateHost %s → 202 status=%s took=%v", name, info.Status, time.Since(start).Truncate(time.Millisecond))
	return &info, nil
}

// GetHost 查询 Host 详情
func (c *APIClient) GetHost(name string) (*HostInfo, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/hosts/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get host: expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var info HostInfo
	if err := json.Unmarshal(dataBytes, &info); err != nil {
		return nil, fmt.Errorf("unmarshal host info: %w", err)
	}
	return &info, nil
}

// ListHosts 列出所有 Host
func (c *APIClient) ListHosts() ([]HostInfo, error) {
	resp, err := c.doRequest("GET", "/api/v1/hosts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list hosts: expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var hosts []HostInfo
	if err := json.Unmarshal(dataBytes, &hosts); err != nil {
		return nil, fmt.Errorf("unmarshal hosts: %w", err)
	}
	return hosts, nil
}

// DeleteHost 删除 Host，打印耗时
func (c *APIClient) DeleteHost(name string) error {
	start := time.Now()
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/hosts/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete host: expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}
	log.Printf("  [api] DeleteHost %s → 200 took=%v", name, time.Since(start).Truncate(time.Millisecond))
	return nil
}

// GetSavedSessions 查询已保存的 Session
func (c *APIClient) GetSavedSessions(hostName string) ([]SessionState, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/nodes/%s/sessions/saved", hostName), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get saved sessions: expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	var sessions []SessionState
	if err := json.Unmarshal(dataBytes, &sessions); err != nil {
		return nil, fmt.Errorf("unmarshal sessions: %w", err)
	}
	return sessions, nil
}

// CreateSession 在指定 Host 上创建 Session，打印耗时
func (c *APIClient) CreateSession(hostName, sessionName, templateID string) error {
	start := time.Now()
	reqBody := map[string]interface{}{
		"name":        sessionName,
		"template_id": templateID,
		"working_dir": "/home/agent/project",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/api/v1/nodes/%s/sessions", hostName), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create session: expected 200/201, got %d: %s", resp.StatusCode, string(respBody))
	}
	log.Printf("  [api] CreateSession host=%s session=%s → %d took=%v", hostName, sessionName, resp.StatusCode, time.Since(start).Truncate(time.Millisecond))
	return nil
}

// WaitForHostStatus 轮询直到 Host 状态匹配，超时返回错误
func (c *APIClient) WaitForHostStatus(name, targetStatus string, timeout time.Duration) (*HostInfo, error) {
	deadline := time.Now().Add(timeout)
	start := time.Now()
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		info, err := c.GetHost(name)
		if err != nil {
			elapsed := time.Since(start).Truncate(time.Second)
			log.Printf("    [wait] host=%s attempt=%d error=%v elapsed=%s", name, attempt, err, elapsed)
			time.Sleep(2 * time.Second)
			continue
		}
		if info.Status == targetStatus {
			elapsed := time.Since(start).Truncate(time.Second)
			log.Printf("    [wait] host=%s reached %s after %s (%d attempts)", name, targetStatus, elapsed, attempt)
			return info, nil
		}
		elapsed := time.Since(start).Truncate(time.Second)
		log.Printf("    [wait] host=%s status=%s (want %s), elapsed=%s", name, info.Status, targetStatus, elapsed)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("wait for host %s status %s: timeout after %v", name, targetStatus, timeout)
}

// WaitForHostCount 轮询直到 Host 数量匹配
func (c *APIClient) WaitForHostCount(expectedCount int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		hosts, err := c.ListHosts()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if len(hosts) == expectedCount {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("wait for host count %d: timeout after %v", expectedCount, timeout)
}

func (c *APIClient) doRequest(method, path string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	return c.client.Do(req)
}
