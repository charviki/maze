package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hertzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
)

func setupTemplateRouter(store *model.TemplateStore) *route.Engine {
	h := NewTemplateHandler(store)
	r := route.NewEngine(hertzconfig.NewOptions(nil))
	r.GET("/api/v1/templates/:id/config", h.GetTemplateConfig)
	r.PUT("/api/v1/templates/:id/config", h.UpdateTemplateConfig)
	return r
}

func TestGetTemplateConfig_ReadsRealGlobalFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("创建 ~/.claude 失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte("{\"hasCompletedOnboarding\":true}"), 0644); err != nil {
		t.Fatalf("写入 ~/.claude.json 失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{\"theme\":\"dark\"}"), 0644); err != nil {
		t.Fatalf("写入 settings.json 失败: %v", err)
	}

	store := model.NewTemplateStore(filepath.Join(t.TempDir(), "templates.json"), logutil.NewNop())
	r := setupTemplateRouter(store)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/templates/claude/config", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)
	resp := parseResponse(w.Body.Bytes())
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("期望 data 为对象")
	}
	files, ok := data["files"].([]interface{})
	if !ok || len(files) == 0 {
		t.Fatalf("期望 files 非空, 实际 %#v", data["files"])
	}
}

func TestUpdateTemplateConfig_ReturnsConflictPayload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("创建 ~/.claude 失败: %v", err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte("{\"theme\":\"dark\"}"), 0644); err != nil {
		t.Fatalf("写入 settings.json 失败: %v", err)
	}

	store := model.NewTemplateStore(filepath.Join(t.TempDir(), "templates.json"), logutil.NewNop())
	r := setupTemplateRouter(store)

	getResp := ut.PerformRequest(r, http.MethodGet, "/api/v1/templates/claude/config", nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("读取 template config 失败: %d", getResp.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(getResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	data := payload["data"].(map[string]interface{})
	files := data["files"].([]interface{})

	var baseHash string
	for _, item := range files {
		fileObj := item.(map[string]interface{})
		if fileObj["path"] == "~/.claude/settings.json" {
			baseHash = fileObj["hash"].(string)
			break
		}
	}
	if baseHash == "" {
		t.Fatal("未找到 ~/.claude/settings.json 的 hash")
	}

	if err := os.WriteFile(settingsPath, []byte("{\"theme\":\"light\"}"), 0644); err != nil {
		t.Fatalf("模拟外部改写失败: %v", err)
	}

	body := fmt.Sprintf(`{"files":[{"path":"~/.claude/settings.json","content":"{\"theme\":\"solarized\"}","base_hash":"%s"}]}`, baseHash)
	w := ut.PerformRequest(r, http.MethodPut, "/api/v1/templates/claude/config", &ut.Body{
		Body: strings.NewReader(body),
		Len:  len(body),
	}, ut.Header{Key: "Content-Type", Value: "application/json"})

	assert.DeepEqual(t, http.StatusConflict, w.Code)
	resp := parseResponse(w.Body.Bytes())
	assert.DeepEqual(t, "config_conflict", resp["code"])
	assert.DeepEqual(t, "配置已变更，请重新加载后再修改", resp["message"])
}
