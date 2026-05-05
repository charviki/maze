package configutil

import "testing"

func TestConfigFileStruct(t *testing.T) {
	cf := ConfigFile{Path: "/tmp/test.txt", Content: "hello"}
	if cf.Path != "/tmp/test.txt" {
		t.Errorf("Path = %q", cf.Path)
	}
	if cf.Content != "hello" {
		t.Errorf("Content = %q", cf.Content)
	}
}

func TestConfigLayerStruct(t *testing.T) {
	layer := ConfigLayer{
		Env: map[string]string{"A": "1"},
		Files: []ConfigFile{
			{Path: "/tmp/a.txt", Content: "a"},
		},
	}
	if len(layer.Env) != 1 {
		t.Errorf("Env length = %d", len(layer.Env))
	}
	if len(layer.Files) != 1 {
		t.Errorf("Files length = %d", len(layer.Files))
	}
}

func TestEnvDefStruct(t *testing.T) {
	def := EnvDef{
		Key:         "API_KEY",
		Label:       "API Key",
		Required:    true,
		Placeholder: "sk-...",
		Sensitive:   true,
	}
	if !def.Sensitive {
		t.Error("Sensitive should be true")
	}
}

func TestFileDefStruct(t *testing.T) {
	def := FileDef{
		Path:           "/etc/config.yaml",
		Label:          "Config",
		Required:       false,
		DefaultContent: "key: value",
	}
	if def.Path != "/etc/config.yaml" {
		t.Errorf("Path = %q", def.Path)
	}
}
