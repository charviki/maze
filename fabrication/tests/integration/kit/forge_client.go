package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ForgeTestClient 是 The Forge 服务的轻量测试客户端。
type ForgeTestClient struct {
	baseURL string
	client  *http.Client
}

// NewForgeTestClient 通过 Director Core 登录获取 JWT，创建带认证的 The Forge 客户端。
func NewForgeTestClient(cfg *TestConfig) (*ForgeTestClient, error) {
	loginResult, err := LoginAdmin(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("login for forge client: %w", err)
	}
	return &ForgeTestClient{
		baseURL: cfg.TheForgeURL,
		client: &http.Client{
			Timeout:   30 * 1000 * 1000 * 1000, // 30s
			Transport: &authTransport{token: loginResult.AccessToken},
		},
	}, nil
}

// authTransport 为请求注入 Authorization header。
type authTransport struct {
	token    string
	delegate http.RoundTripper
}

// RoundTrip 注入 Bearer token 后转发请求。
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	delegate := t.delegate
	if delegate == nil {
		delegate = http.DefaultTransport
	}
	return delegate.RoundTrip(req)
}

// --- Helpers ---

func (c *ForgeTestClient) doJSON(method, path string, body any) (int, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func (c *ForgeTestClient) get(path string) (int, []byte, error) {
	return c.doJSON(http.MethodGet, path, nil)
}

func (c *ForgeTestClient) post(path string, body any) (int, []byte, error) {
	return c.doJSON(http.MethodPost, path, body)
}

func (c *ForgeTestClient) put(path string, body any) (int, []byte, error) {
	return c.doJSON(http.MethodPut, path, body)
}

func (c *ForgeTestClient) delete(path string) (int, []byte, error) {
	return c.doJSON(http.MethodDelete, path, nil)
}

// --- Health ---

// HealthResponse 表示健康检查响应。
type HealthResponse struct {
	Status string `json:"status"`
}

// GetHealth 调用 /health 端点。
func (c *ForgeTestClient) GetHealth() (*HealthResponse, error) {
	status, body, err := c.get("/health")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("health status=%d body=%s", status, body)
	}
	var resp HealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Archive ---

// ForgeArchive 表示知识库 Archive。
type ForgeArchive struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Author      string `json:"author"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateArchive 创建 Archive。
func (c *ForgeTestClient) CreateArchive(name, description, icon string) (*ForgeArchive, error) {
	status, body, err := c.post("/api/v1/archives", map[string]string{
		"name": name, "description": description, "icon": icon,
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("create archive status=%d body=%s", status, body)
	}
	var archive ForgeArchive
	if err := json.Unmarshal(body, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

// GetArchive 获取指定 Archive。
func (c *ForgeTestClient) GetArchive(id string) (*ForgeArchive, error) {
	status, body, err := c.get("/api/v1/archives/" + id)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get archive status=%d body=%s", status, body)
	}
	var archive ForgeArchive
	if err := json.Unmarshal(body, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

// ListArchives 列出所有 Archives。
func (c *ForgeTestClient) ListArchives() ([]ForgeArchive, error) {
	status, body, err := c.get("/api/v1/archives")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("list archives status=%d body=%s", status, body)
	}
	var resp struct {
		Archives []ForgeArchive `json:"archives"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Archives, nil
}

// UpdateArchive 更新指定 Archive。
func (c *ForgeTestClient) UpdateArchive(id, name, description, icon string) (*ForgeArchive, error) {
	status, body, err := c.put("/api/v1/archives/"+id, map[string]string{
		"name": name, "description": description, "icon": icon,
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("update archive status=%d body=%s", status, body)
	}
	var archive ForgeArchive
	if err := json.Unmarshal(body, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

// DeleteArchive 删除指定 Archive。
func (c *ForgeTestClient) DeleteArchive(id string) error {
	status, _, err := c.delete("/api/v1/archives/" + id)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("delete archive status=%d", status)
	}
	return nil
}

// --- Doc ---

// ForgeDoc 表示知识库 Doc。
type ForgeDoc struct {
	ID          string            `json:"id"`
	ArchiveID   string            `json:"archiveId"`
	ParentID    string            `json:"parentId,omitempty"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Summary     string            `json:"summary,omitempty"`
	Status      string            `json:"status,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Assignee    string            `json:"assignee,omitempty"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Visibility  string            `json:"visibility"`
	SharedWith  []string          `json:"sharedWith,omitempty"`
	Attachments []ForgeAttachment `json:"attachments,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

// ForgeAttachment 表示附件。
type ForgeAttachment struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// ListDocsResponse 表示列表 Docs 响应。
type ListDocsResponse struct {
	Items []ForgeDoc `json:"items"`
	Total int        `json:"total"`
}

// CreateDoc 创建 Doc（简版）。
func (c *ForgeTestClient) CreateDoc(archiveID, title, content string) (*ForgeDoc, error) {
	status, body, err := c.post("/api/v1/docs", map[string]any{
		"archiveId": archiveID, "title": title, "content": content,
		"visibility": "public",
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("create doc status=%d body=%s", status, body)
	}
	var doc ForgeDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// CreateDocFull 创建 Doc（完整参数）。
func (c *ForgeTestClient) CreateDocFull(archiveID, parentID, title, content string, status *string, priority *string, assignee string, tags []string, visibility string) (*ForgeDoc, error) {
	payload := map[string]any{
		"archiveId": archiveID, "title": title, "content": content,
		"visibility": visibility, "tags": tags, "assignee": assignee,
	}
	if parentID != "" {
		payload["parentId"] = parentID
	}
	if status != nil {
		payload["status"] = *status
	}
	if priority != nil {
		payload["priority"] = *priority
	}
	code, body, err := c.post("/api/v1/docs", payload)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return nil, fmt.Errorf("create doc status=%d body=%s", code, body)
	}
	var doc ForgeDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetDoc 获取指定 Doc。
func (c *ForgeTestClient) GetDoc(id string) (*ForgeDoc, error) {
	status, body, err := c.get("/api/v1/docs/" + id)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get doc status=%d body=%s", status, body)
	}
	var doc ForgeDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocs 列出 Docs（支持 query params）。
func (c *ForgeTestClient) ListDocs(params string) (*ListDocsResponse, error) {
	path := "/api/v1/docs"
	if params != "" {
		path += "?" + params
	}
	status, body, err := c.get(path)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("list docs status=%d body=%s", status, body)
	}
	var resp ListDocsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateDoc 更新指定 Doc 字段。
func (c *ForgeTestClient) UpdateDoc(id string, updates map[string]any) (*ForgeDoc, error) {
	status, body, err := c.put("/api/v1/docs/"+id, updates)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("update doc status=%d body=%s", status, body)
	}
	var doc ForgeDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// DeleteDoc 删除指定 Doc。
func (c *ForgeTestClient) DeleteDoc(id string) error {
	status, _, err := c.delete("/api/v1/docs/" + id)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("delete doc status=%d", status)
	}
	return nil
}

// SearchDocs 搜索 Docs。
func (c *ForgeTestClient) SearchDocs(query string, archiveID string) ([]ForgeDoc, error) {
	path := "/api/v1/docs:search?q=" + query
	if archiveID != "" {
		path += "&archiveId=" + archiveID
	}
	status, body, err := c.get(path)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("search docs status=%d body=%s", status, body)
	}
	var resp struct {
		Items []ForgeDoc `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// DocTreeNode 表示 Doc 树节点。
type DocTreeNode struct {
	Doc      ForgeDoc        `json:"doc"`
	Children []DocTreeNode   `json:"children,omitempty"`
}

// GetDocTree 获取 Doc 树形结构。
func (c *ForgeTestClient) GetDocTree(params string) ([]DocTreeNode, error) {
	path := "/api/v1/docs:tree"
	if params != "" {
		path += "?" + params
	}
	status, body, err := c.get(path)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get doc tree status=%d body=%s", status, body)
	}
	var resp struct {
		Nodes []DocTreeNode `json:"nodes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}

// GetDocAncestors 获取指定 Doc 的祖先链。
func (c *ForgeTestClient) GetDocAncestors(id string) ([]ForgeDoc, error) {
	status, body, err := c.get("/api/v1/docs/" + id + "/ancestors")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get ancestors status=%d body=%s", status, body)
	}
	var resp struct {
		Ancestors []ForgeDoc `json:"ancestors"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Ancestors, nil
}

// --- Links ---

// ForgeLink 表示两个 Doc 之间的关联。
type ForgeLink struct {
	ID           string `json:"id"`
	SourceID     string `json:"sourceId"`
	TargetID     string `json:"targetId"`
	RelationType string `json:"relationType"`
	SourceTitle  string `json:"sourceTitle,omitempty"`
	TargetTitle  string `json:"targetTitle,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

// CreateLink 在两个 Doc 之间创建关联。
func (c *ForgeTestClient) CreateLink(docID, targetID, relationType string) (*ForgeLink, error) {
	status, body, err := c.post("/api/v1/docs/"+docID+"/links", map[string]string{
		"targetId": targetID, "relationType": relationType,
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("create link status=%d body=%s", status, body)
	}
	var link ForgeLink
	if err := json.Unmarshal(body, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// GetLinks 获取指定 Doc 的所有关联。
func (c *ForgeTestClient) GetLinks(docID string) ([]ForgeLink, error) {
	status, body, err := c.get("/api/v1/docs/" + docID + "/links")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get links status=%d body=%s", status, body)
	}
	var resp struct {
		Links []ForgeLink `json:"links"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Links, nil
}

// DeleteLink 删除指定关联。
func (c *ForgeTestClient) DeleteLink(docID, linkID string) error {
	status, _, err := c.delete("/api/v1/docs/" + docID + "/links/" + linkID)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("delete link status=%d", status)
	}
	return nil
}

// --- Unauthenticated request ---

// GetUnauthenticated 使用无认证的裸 client 发送 GET 请求，返回状态码。
func (c *ForgeTestClient) GetUnauthenticated(path string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

// BuildQueryParams 从 key-value 对构建 URL query string。
func BuildQueryParams(params map[string]string) string {
	var parts []string
	for k, v := range params {
		if v != "" {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, "&")
}
