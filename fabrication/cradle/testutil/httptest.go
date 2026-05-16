package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// PostJSON sends a POST request with JSON body and decodes the response.
func PostJSON[T any](t *testing.T, client *http.Client, url string, payload any) T {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("post %s status = %d, body = %s", url, resp.StatusCode, string(raw))
	}
	return DecodeJSON[T](t, resp.Body)
}

// GetJSON sends a GET request and decodes the JSON response.
func GetJSON[T any](t *testing.T, client *http.Client, url string) T {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("get %s status = %d, body = %s", url, resp.StatusCode, string(raw))
	}
	return DecodeJSON[T](t, resp.Body)
}

// DecodeJSON decodes a JSON response body into T.
func DecodeJSON[T any](t *testing.T, body io.Reader) T {
	t.Helper()

	var result T
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return result
}

// AnySlice type-asserts value to []any.
func AnySlice(value any) []any {
	items, _ := value.([]any)
	return items
}
