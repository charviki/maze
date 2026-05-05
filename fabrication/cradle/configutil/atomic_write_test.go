package configutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.txt")
	data := []byte("hello world")

	if err := AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("content = %q, want %q", string(got), "hello world")
	}
}

func TestAtomicWriteFile_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "executable.sh")
	data := []byte("#!/bin/sh\necho hi")

	if err := AtomicWriteFile(path, data, 0755); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("permissions = %o, want 0755", info.Mode().Perm())
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.txt")

	if err := AtomicWriteFile(path, []byte("first"), 0644); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := AtomicWriteFile(path, []byte("second"), 0644); err != nil {
		t.Fatalf("second write: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "second" {
		t.Errorf("content = %q, want second", string(got))
	}
}

func TestAtomicWriteFile_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	if err := AtomicWriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("AtomicWriteFile with empty data: %v", err)
	}

	got, _ := os.ReadFile(path)
	if len(got) != 0 {
		t.Errorf("file should be empty, got %q", string(got))
	}
}

func TestAtomicWriteFile_InvalidDir(t *testing.T) {
	path := "/nonexistent/dir/target.txt"
	err := AtomicWriteFile(path, []byte("data"), 0644)
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}
