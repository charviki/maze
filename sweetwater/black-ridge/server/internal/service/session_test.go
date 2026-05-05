package service

import "testing"

func TestSessionStruct(t *testing.T) {
	s := Session{
		ID:          "sess-1",
		Name:        "my-session",
		Status:      "running",
		CreatedAt:   "2024-01-01T00:00:00Z",
		WindowCount: 3,
	}
	if s.ID != "sess-1" {
		t.Errorf("ID = %q, want sess-1", s.ID)
	}
	if s.Name != "my-session" {
		t.Errorf("Name = %q, want my-session", s.Name)
	}
	if s.Status != "running" {
		t.Errorf("Status = %q, want running", s.Status)
	}
	if s.WindowCount != 3 {
		t.Errorf("WindowCount = %d, want 3", s.WindowCount)
	}
}

func TestConfigItemStruct(t *testing.T) {
	c := ConfigItem{Type: "env", Key: "PATH", Value: "/usr/bin"}
	if c.Type != "env" {
		t.Errorf("Type = %q, want env", c.Type)
	}
	if c.Key != "PATH" {
		t.Errorf("Key = %q, want PATH", c.Key)
	}
}

func TestConfigScope_Constants(t *testing.T) {
	if ConfigScopeGlobal != "global" {
		t.Errorf("ConfigScopeGlobal = %q, want global", ConfigScopeGlobal)
	}
	if ConfigScopeProject != "project" {
		t.Errorf("ConfigScopeProject = %q, want project", ConfigScopeProject)
	}
}

func TestSaveConfigRequest(t *testing.T) {
	req := SaveConfigRequest{
		Files: []ConfigFileUpdate{
			{Path: "/etc/config.yaml", Content: "key: value", BaseHash: "abc123"},
		},
	}
	if len(req.Files) != 1 {
		t.Errorf("Files length = %d, want 1", len(req.Files))
	}
	if req.Files[0].BaseHash != "abc123" {
		t.Errorf("BaseHash = %q, want abc123", req.Files[0].BaseHash)
	}
}

func TestConfigConflict(t *testing.T) {
	cc := ConfigConflict{Path: "/etc/config.yaml", CurrentHash: "def456"}
	if cc.Path != "/etc/config.yaml" {
		t.Errorf("Path = %q", cc.Path)
	}
	if cc.CurrentHash != "def456" {
		t.Errorf("CurrentHash = %q", cc.CurrentHash)
	}
}
