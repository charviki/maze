package storeutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
)

type testData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestNewJSONStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "default", Count: 0}, logutil.NewNop())
	if store == nil {
		t.Fatal("NewJSONStore returned nil")
	}

	got := store.Get()
	if got.Name != "default" {
		t.Errorf("Name = %q, want %q", got.Name, "default")
	}
}

func TestJSONStore_Get(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "test", Count: 42}, logutil.NewNop())

	got := store.Get()
	if got.Name != "test" {
		t.Errorf("Name = %q, want %q", got.Name, "test")
	}
	if got.Count != 42 {
		t.Errorf("Count = %d, want 42", got.Count)
	}
}

func TestJSONStore_Update(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "initial", Count: 0}, logutil.NewNop())

	err := store.Update(func(data *testData) {
		data.Name = "updated"
		data.Count = 100
	}, true)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got := store.Get()
	if got.Name != "updated" {
		t.Errorf("Name = %q, want %q", got.Name, "updated")
	}
	if got.Count != 100 {
		t.Errorf("Count = %d, want 100", got.Count)
	}

	exists, _ := os.Stat(path)
	if exists == nil {
		t.Fatal("persisted file should exist")
	}
}

func TestJSONStore_Update_NoPersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "initial", Count: 0}, logutil.NewNop())

	err := store.Update(func(data *testData) {
		data.Count = 99
	}, false)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got := store.Get()
	if got.Count != 99 {
		t.Errorf("Count = %d, want 99", got.Count)
	}
}

func TestJSONStore_Save(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "initial", Count: 0}, logutil.NewNop())

	_ = store.Update(func(data *testData) {
		data.Count = 5
	}, false)

	err := store.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted file failed: %v", err)
	}

	var persisted testData
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("unmarshal persisted file failed: %v", err)
	}
	if persisted.Count != 5 {
		t.Errorf("persisted Count = %d, want 5", persisted.Count)
	}
}

func TestJSONStore_View(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "test", Count: 10}, logutil.NewNop())

	var seen testData
	store.View(func(data *testData) {
		seen = *data
	})

	if seen.Name != "test" {
		t.Errorf("Name = %q, want %q", seen.Name, "test")
	}
	if seen.Count != 10 {
		t.Errorf("Count = %d, want 10", seen.Count)
	}
}

func TestJSONStore_PersistAndRecover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store1 := NewJSONStore(path, testData{Name: "initial", Count: 0}, logutil.NewNop())
	_ = store1.Update(func(data *testData) {
		data.Name = "persisted"
		data.Count = 777
	}, true)

	store2 := NewJSONStore(path, testData{Name: "default", Count: 0}, logutil.NewNop())
	got := store2.Get()
	if got.Name != "persisted" {
		t.Errorf("Name = %q, want %q", got.Name, "persisted")
	}
	if got.Count != 777 {
		t.Errorf("Count = %d, want 777", got.Count)
	}
}

func TestJSONStore_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	store := NewJSONStore(path, testData{Name: "concurrent", Count: 0}, logutil.NewNop())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Update(func(data *testData) {
				data.Count++
			}, false)
		}()
	}
	wg.Wait()

	got := store.Get()
	if got.Count != 100 {
		t.Errorf("Count = %d, want 100", got.Count)
	}
}

func TestJSONStore_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	if err := os.WriteFile(path, []byte("{invalid!!!}"), 0644); err != nil {
		t.Fatalf("write corrupted file: %v", err)
	}

	store := NewJSONStore(path, testData{Name: "fallback", Count: -1}, logutil.NewNop())
	got := store.Get()
	if got.Name != "fallback" {
		t.Errorf("Name = %q, want fallback (initial value)", got.Name)
	}
	if got.Count != -1 {
		t.Errorf("Count = %d, want -1 (initial value)", got.Count)
	}

	_ = store.Save()
	if _, err := os.Stat(path); err != nil {
		t.Error("should be able to create valid file after corruption")
	}
}
