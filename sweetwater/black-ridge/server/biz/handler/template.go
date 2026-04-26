package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// 模板管理 handler
type TemplateHandler struct {
	store       *model.TemplateStore
	configFiles *service.ConfigFileService
}

// 创建 TemplateHandler
func NewTemplateHandler(store *model.TemplateStore) *TemplateHandler {
	return &TemplateHandler{
		store:       store,
		configFiles: service.NewConfigFileService(),
	}
}

// 列出所有模板
func (h *TemplateHandler) ListTemplates(_ context.Context, c *app.RequestContext) {
	templates := h.store.List()
	httputil.Success(c, templates)
}

// 获取指定模板
func (h *TemplateHandler) GetTemplate(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	tpl := h.store.Get(id)
	if tpl == nil {
		httputil.Error(c, http.StatusNotFound, "template not found")
		return
	}
	httputil.Success(c, tpl)
}

// GetTemplateConfig 返回模板声明的全局固定文件在节点上的真实内容。
func (h *TemplateHandler) GetTemplateConfig(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	tpl := h.store.Get(id)
	if tpl == nil {
		httputil.Error(c, http.StatusNotFound, "template not found")
		return
	}

	files, err := h.configFiles.ReadGlobalFiles(tpl.Defaults.Files)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, model.TemplateConfigView{
		TemplateID: tpl.ID,
		Scope:      model.ConfigScopeGlobal,
		Files:      files,
	})
}

// UpdateTemplateConfig 以真实全局文件为目标保存模板配置，并同步模板中的内容定义。
func (h *TemplateHandler) UpdateTemplateConfig(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var req model.SaveConfigRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	existing := h.store.Get(id)
	if existing == nil {
		httputil.Error(c, http.StatusNotFound, "template not found")
		return
	}

	files, err := h.configFiles.SaveGlobalFiles(existing.Defaults.Files, req.Files)
	if err != nil {
		var conflictErr *service.ConfigConflictError
		if errors.As(err, &conflictErr) {
			c.JSON(http.StatusConflict, map[string]interface{}{
				"status":    "error",
				"code":      "config_conflict",
				"message":   conflictErr.Error(),
				"conflicts": conflictErr.Conflicts,
			})
			return
		}
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	updatedTemplate := cloneTemplateWithGlobalContents(existing, req.Files)
	if err := h.store.Set(updatedTemplate); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, model.TemplateConfigView{
		TemplateID: updatedTemplate.ID,
		Scope:      model.ConfigScopeGlobal,
		Files:      files,
	})
}

// 创建新模板（自动设置 builtin=false）
func (h *TemplateHandler) CreateTemplate(_ context.Context, c *app.RequestContext) {
	var tpl model.SessionTemplate
	if err := c.Bind(&tpl); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if tpl.ID == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}
	if tpl.Name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	if h.store.Get(tpl.ID) != nil {
		httputil.Error(c, http.StatusConflict, fmt.Sprintf("template %s already exists", tpl.ID))
		return
	}

	tpl.Builtin = false
	if err := h.store.Set(&tpl); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, tpl)
}

// 更新模板（保留 builtin 标记）
func (h *TemplateHandler) UpdateTemplate(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var tpl model.SessionTemplate
	if err := c.Bind(&tpl); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	existing := h.store.Get(id)
	if existing == nil {
		httputil.Error(c, http.StatusNotFound, "template not found")
		return
	}

	tpl.ID = id
	tpl.Builtin = existing.Builtin
	// 固定路径配置改走 /templates/:id/config，旧的整模板更新接口只保留模板元信息修改能力，
	// 避免路径集合和文件内容再次被任意请求体覆盖。
	tpl.Defaults.Files = cloneConfigFiles(existing.Defaults.Files)
	tpl.SessionSchema.FileDefs = cloneFileDefs(existing.SessionSchema.FileDefs)
	if err := h.store.Set(&tpl); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, tpl)
}

// 删除模板（禁止删除内置模板）
func (h *TemplateHandler) DeleteTemplate(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	existing := h.store.Get(id)
	if existing == nil {
		httputil.Error(c, http.StatusNotFound, "template not found")
		return
	}
	if existing.Builtin {
		httputil.Error(c, http.StatusForbidden, "cannot delete built-in template")
		return
	}

	if err := h.store.Delete(id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.Success(c, nil)
}

func cloneTemplateWithGlobalContents(existing *model.SessionTemplate, updates []model.ConfigFileUpdate) *model.SessionTemplate {
	cloned := *existing
	cloned.Defaults.Env = cloneStringMap(existing.Defaults.Env)
	cloned.Defaults.Files = cloneConfigFiles(existing.Defaults.Files)
	cloned.SessionSchema.EnvDefs = append([]model.EnvDef(nil), existing.SessionSchema.EnvDefs...)
	cloned.SessionSchema.FileDefs = cloneFileDefs(existing.SessionSchema.FileDefs)

	contentByPath := make(map[string]string, len(updates))
	for _, update := range updates {
		contentByPath[update.Path] = update.Content
	}

	// 模板存储只同步内容，不改固定路径集合，避免旧的整模板更新接口把路径模型重新放开。
	for i := range cloned.Defaults.Files {
		if content, ok := contentByPath[cloned.Defaults.Files[i].Path]; ok {
			cloned.Defaults.Files[i].Content = content
		}
	}

	return &cloned
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneConfigFiles(src []model.ConfigFile) []model.ConfigFile {
	if src == nil {
		return nil
	}
	dst := make([]model.ConfigFile, len(src))
	copy(dst, src)
	return dst
}

func cloneFileDefs(src []model.FileDef) []model.FileDef {
	if src == nil {
		return nil
	}
	dst := make([]model.FileDef, len(src))
	copy(dst, src)
	return dst
}
