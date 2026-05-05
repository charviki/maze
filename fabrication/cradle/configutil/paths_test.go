package configutil

import "testing"

func TestExpandHomePath_NotTilde(t *testing.T) {
	abs := "/absolute/path"
	result := ExpandHomePath(abs)
	if result != abs {
		t.Errorf("ExpandHomePath(%q) = %q, want %q", abs, result, abs)
	}
}

func TestExpandHomePath_WithTildePrefix(t *testing.T) {
	result := ExpandHomePath("~/data")
	if result == "" {
		t.Error("ExpandHomePath should not return empty")
	}
}

func TestExpandHomePath_RelativePath(t *testing.T) {
	result := ExpandHomePath("config.yaml")
	if result != "config.yaml" {
		t.Errorf("ExpandHomePath = %q, want config.yaml", result)
	}
}

func TestExpandHomePath_Empty(t *testing.T) {
	result := ExpandHomePath("")
	if result != "" {
		t.Errorf("ExpandHomePath = %q, want empty", result)
	}
}
