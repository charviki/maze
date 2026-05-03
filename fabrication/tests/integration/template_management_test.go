//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestTemplateCRUD — Given: 已上线的 Host; When: 创建→查询→更新→删除模板; Then: 全生命周期正确
func TestTemplateCRUD(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] creating template...")
	tmplID := "test-template-1"
	tmplName := "test-template"
	command := "bash"
	createBody := client.TemplateServiceCreateTemplateBody{
		Template: &client.V1SessionTemplate{
			Id:      &tmplID,
			Name:    &tmplName,
			Command: &command,
		},
	}
	created, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceCreateTemplate(context.Background(), nodeName).
		Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create template failed: %v", err)
	}
	tmplID = created.GetId()
	t.Logf("[step] created template: id=%s name=%s", tmplID, created.GetName())

	t.Log("[step] getting template...")
	tmpl, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceGetTemplate(context.Background(), nodeName, tmplID).Execute()
	if err != nil {
		t.Fatalf("get template failed: %v", err)
	}
	if tmpl.GetName() != tmplName {
		t.Errorf("expected name=%s, got=%s", tmplName, tmpl.GetName())
	}

	t.Log("[step] listing templates...")
	listResp, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceListTemplates(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("list templates failed: %v", err)
	}
	found := false
	for _, t := range listResp.GetTemplates() {
		if t.GetId() == tmplID {
			found = true
		}
	}
	if !found {
		t.Errorf("template %s not found in list", tmplID)
	}

	t.Log("[step] updating template...")
	newName := "updated-template"
	updateBody := client.TemplateServiceUpdateTemplateBody{
		Template: &client.V1SessionTemplate{
			Name: &newName,
		},
	}
	updated, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceUpdateTemplate(context.Background(), nodeName, tmplID).
		Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update template failed: %v", err)
	}
	if updated.GetName() != newName {
		t.Errorf("expected updated name=%s, got=%s", newName, updated.GetName())
	}

	t.Log("[step] deleting template...")
	_, _, err = h.apiClient.TemplateServiceAPI.TemplateServiceDeleteTemplate(context.Background(), nodeName, tmplID).Execute()
	if err != nil {
		t.Fatalf("delete template failed: %v", err)
	}

	t.Log("[step] PASS: template CRUD completed")
}

// TestTemplateConfig — Given: 已上线的 Host 和已创建的模板; When: 查询/更新模板配置; Then: 配置正确返回
func TestTemplateConfig(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] creating template for config test...")
	tmplIDCfg := "config-test-template-1"
	tmplName := "config-test-template"
	command := "bash"
	configPath := "~/.config/test-template.yaml"
	configLabel := "Test Template Config"
	defaultContent := "mode: template\n"
	createBody := client.TemplateServiceCreateTemplateBody{
		Template: &client.V1SessionTemplate{
			Id:      &tmplIDCfg,
			Name:    &tmplName,
			Command: &command,
			Defaults: &client.V1ConfigLayer{
				Files: []client.V1ConfigFile{
					{
						Path:    ptrStr(configPath),
						Content: ptrStr(defaultContent),
					},
				},
			},
			SessionSchema: &client.V1SessionSchema{
				FileDefs: []client.V1FileDef{
					{
						Path:           ptrStr(configPath),
						Label:          ptrStr(configLabel),
						DefaultContent: ptrStr(defaultContent),
					},
				},
			},
		},
	}
	created, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceCreateTemplate(context.Background(), nodeName).
		Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create template failed: %v", err)
	}
	tmplID := created.GetId()

	t.Log("[step] getting template config...")
	configView, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceGetTemplateConfig(context.Background(), nodeName, tmplID).Execute()
	if err != nil {
		t.Fatalf("get template config failed: %v", err)
	}
	if len(configView.GetFiles()) == 0 {
		t.Fatal("expected template config to expose editable files")
	}
	firstFile := configView.GetFiles()[0]
	updatedContent := firstFile.GetContent() + "\n# integration template config\n"
	t.Logf("[step] template config: template_id=%s scope=%s", configView.GetTemplateId(), configView.GetScope())

	t.Log("[step] updating template config...")
	updateBody := client.TemplateServiceUpdateTemplateConfigBody{
		Files: []client.V1ConfigFileUpdate{
			{
				Path:     ptrStr(firstFile.GetPath()),
				Content:  ptrStr(updatedContent),
				BaseHash: ptrStr(firstFile.GetHash()),
			},
		},
	}
	updatedView, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceUpdateTemplateConfig(context.Background(), nodeName, tmplID).
		Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update template config failed: %v", err)
	}
	t.Logf("[step] updated config: template_id=%s scope=%s", updatedView.GetTemplateId(), updatedView.GetScope())
	if len(updatedView.GetFiles()) == 0 || updatedView.GetFiles()[0].GetContent() != updatedContent {
		t.Fatal("expected template config update to return persisted content")
	}

	t.Log("[step] re-reading template config...")
	rereadView, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceGetTemplateConfig(context.Background(), nodeName, tmplID).Execute()
	if err != nil {
		t.Fatalf("re-read template config failed: %v", err)
	}
	if len(rereadView.GetFiles()) == 0 || rereadView.GetFiles()[0].GetContent() != updatedContent {
		t.Fatal("expected template config re-read to return persisted content")
	}

	t.Log("[step] PASS: template config get/update succeeded")
}
