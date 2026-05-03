package httputil

import (
	"net"
	"net/http"
	"testing"
)

func TestValidateTargetURL(t *testing.T) {
	testCases := []struct {
		name         string
		rawURL       string
		allowPrivate bool
		wantErr      bool
	}{
		{name: "public ip allowed", rawURL: "https://8.8.8.8/api", wantErr: false},
		{name: "private ip blocked", rawURL: "http://127.0.0.1:8080", wantErr: true},
		{name: "private ip allowed explicitly", rawURL: "ws://127.0.0.1:8080/ws", allowPrivate: true, wantErr: false},
		{name: "unsupported scheme", rawURL: "ftp://8.8.8.8/file", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTargetURL(tc.rawURL, tc.allowPrivate)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateTargetURL() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	testCases := []struct {
		addr string
		want bool
	}{
		{addr: "127.0.0.1", want: true},
		{addr: "10.1.2.3", want: true},
		{addr: "172.16.2.3", want: true},
		{addr: "192.168.1.1", want: true},
		{addr: "169.254.1.1", want: true},
		{addr: "::1", want: true},
		{addr: "fc00::1", want: true},
		{addr: "8.8.8.8", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.addr, func(t *testing.T) {
			ip := netParseIP(t, tc.addr)
			if got := IsPrivateIP(ip); got != tc.want {
				t.Fatalf("IsPrivateIP(%s) = %v, 期望 %v", tc.addr, got, tc.want)
			}
		})
	}
}

func TestNewUpgrader(t *testing.T) {
	upgrader := NewUpgrader([]string{"http://allowed.test"})
	if upgrader == nil {
		t.Fatal("NewUpgrader() = nil, 期望非 nil")
	}

	allowedReq, _ := http.NewRequest("GET", "http://example.test", nil)
	allowedReq.Header.Set("Origin", "http://allowed.test")
	if !upgrader.CheckOrigin(allowedReq) {
		t.Fatal("CheckOrigin() 对允许来源返回 false")
	}

	blockedReq, _ := http.NewRequest("GET", "http://example.test", nil)
	blockedReq.Header.Set("Origin", "http://blocked.test")
	if upgrader.CheckOrigin(blockedReq) {
		t.Fatal("CheckOrigin() 对未授权来源返回 true")
	}
}

func netParseIP(t *testing.T, raw string) net.IP {
	t.Helper()
	ip := net.ParseIP(raw)
	if ip == nil {
		t.Fatalf("ParseIP(%q) = nil", raw)
	}
	return ip
}
