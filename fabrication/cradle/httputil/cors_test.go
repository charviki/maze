package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowsAll(t *testing.T) {
	handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		method string
		origin string
	}{
		{name: "GET / no origin", method: "GET", origin: ""},
		{name: "GET with origin", method: "GET", origin: "http://example.com"},
		{name: "OPTIONS preflight", method: "OPTIONS", origin: "http://example.com"},
		{name: "POST", method: "POST", origin: "http://app.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.method == "OPTIONS" {
				if rec.Code != http.StatusNoContent {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
				}
			} else {
				if rec.Code != http.StatusOK {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
				}
			}
			if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want *", allowOrigin)
			}
		})
	}
}

func TestCORSWithOrigins_Restricted(t *testing.T) {
	handler := CORSWithOrigins([]string{"http://myapp.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	tests := []struct {
		name       string
		origin     string
		method     string
		wantStatus int
		wantOrigin string
	}{
		{
			name:       "allowed origin",
			origin:     "http://myapp.com",
			method:     "GET",
			wantStatus: http.StatusOK,
			wantOrigin: "http://myapp.com",
		},
		{
			name:       "OPTIONS forbidden origin",
			origin:     "http://evil.com",
			method:     "OPTIONS",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "GET with forbidden origin passes (browser skips CORS for simple requests)",
			origin:     "http://evil.com",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
		{
			name:       "no origin header",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantOrigin != "" {
				if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
					t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
				}
			}
		})
	}
}

func TestCORSWithOrigins_EmptyIsAll(t *testing.T) {
	handler := CORSWithOrigins(nil)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://any.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("should allow all when origins is empty")
	}
}
