package configutil

import (
	"testing"
)

func TestMergeConfigLayers_Empty(t *testing.T) {
	result := MergeConfigLayers()
	if len(result.Env) != 0 {
		t.Errorf("Env length = %d, want 0", len(result.Env))
	}
	if len(result.Files) != 0 {
		t.Errorf("Files length = %d, want 0", len(result.Files))
	}
}

func TestMergeConfigLayers_SingleLayer(t *testing.T) {
	layer := ConfigLayer{
		Env:   map[string]string{"KEY": "val"},
		Files: []ConfigFile{{Path: "/tmp/a.txt", Content: "hello"}},
	}
	result := MergeConfigLayers(layer)
	if result.Env["KEY"] != "val" {
		t.Errorf("Env[KEY] = %q, want val", result.Env["KEY"])
	}
	if len(result.Files) != 1 || result.Files[0].Content != "hello" {
		t.Errorf("Files = %+v", result.Files)
	}
}

func TestMergeConfigLayers_Override(t *testing.T) {
	layer1 := ConfigLayer{
		Env:   map[string]string{"KEY": "v1"},
		Files: []ConfigFile{{Path: "/tmp/a.txt", Content: "c1"}},
	}
	layer2 := ConfigLayer{
		Env:   map[string]string{"KEY": "v2"},
		Files: []ConfigFile{{Path: "/tmp/a.txt", Content: "c2"}},
	}
	result := MergeConfigLayers(layer1, layer2)
	if result.Env["KEY"] != "v2" {
		t.Errorf("Env[KEY] = %q, want v2 (latter overrides)", result.Env["KEY"])
	}
	if len(result.Files) != 1 || result.Files[0].Content != "c2" {
		t.Errorf("Files[0].Content = %q, want c2", result.Files[0].Content)
	}
}

func TestMergeConfigLayers_SkipEmptyValues(t *testing.T) {
	layer := ConfigLayer{
		Env:   map[string]string{"KEY": ""},
		Files: []ConfigFile{{Path: "/tmp/a.txt", Content: ""}},
	}
	result := MergeConfigLayers(layer)
	if _, ok := result.Env["KEY"]; ok {
		t.Error("empty env value should be skipped")
	}
	if len(result.Files) != 0 {
		t.Error("empty file content should be skipped")
	}
}

func TestMergeConfigLayers_AddNewKeys(t *testing.T) {
	layer1 := ConfigLayer{Env: map[string]string{"A": "1"}}
	layer2 := ConfigLayer{Env: map[string]string{"B": "2"}}
	result := MergeConfigLayers(layer1, layer2)
	if result.Env["A"] != "1" || result.Env["B"] != "2" {
		t.Errorf("Env = %v, want {A:1, B:2}", result.Env)
	}
}

func TestMergeConfigLayers_DistinctFiles(t *testing.T) {
	layer1 := ConfigLayer{Files: []ConfigFile{{Path: "/tmp/a.txt", Content: "a"}}}
	layer2 := ConfigLayer{Files: []ConfigFile{{Path: "/tmp/b.txt", Content: "b"}}}
	result := MergeConfigLayers(layer1, layer2)
	if len(result.Files) != 2 {
		t.Errorf("Files length = %d, want 2", len(result.Files))
	}
}
