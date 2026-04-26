package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateAgentURL_PrivateIPs(t *testing.T) {
	privateURLs := []struct {
		name string
		url  string
	}{
		{"loopback", "http://127.0.0.1:8080"},
		{"10.x", "http://10.0.0.1:8080"},
		{"172.16.x", "http://172.16.0.1:8080"},
		{"172.31.x", "http://172.31.255.1:8080"},
		{"192.168.x", "http://192.168.1.1:8080"},
		{"169.254.x", "http://169.254.1.1:8080"},
		{"ipv6 loopback", "http://[::1]:8080"},
	}

	for _, tc := range privateURLs {
		t.Run(tc.name, func(t *testing.T) {
			// 先解析域名，确保 DNS 解析到内网 IP
			err := validateAgentURL(tc.url, false)
			if err == nil {
				t.Errorf("期望 %s 被拦截（内网 IP），但通过了校验", tc.url)
			}
		})
	}
}

func TestValidateAgentURL_PublicIPs(t *testing.T) {
	// 使用真实公网 DNS 来测试（需要网络连接）
	publicURLs := []struct {
		name string
		url  string
	}{
		{"google dns", "http://8.8.8.8:8080"},
		{"cloudflare dns", "http://1.1.1.1:8080"},
	}

	for _, tc := range publicURLs {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAgentURL(tc.url, false)
			if err != nil {
				t.Errorf("期望 %s 放行（公网 IP），但被拦截: %v", tc.url, err)
			}
		})
	}
}

func TestValidateAgentURL_AllowPrivate(t *testing.T) {
	// allowPrivate=true 时内网 IP 应放行
	err := validateAgentURL("http://127.0.0.1:8080", true)
	if err != nil {
		t.Errorf("allowPrivate=true 时应放行内网 IP，但被拦截: %v", err)
	}

	err = validateAgentURL("http://10.0.0.1:8080", true)
	if err != nil {
		t.Errorf("allowPrivate=true 时应放行内网 IP，但被拦截: %v", err)
	}
}

func TestValidateAgentURL_UnsupportedScheme(t *testing.T) {
	err := validateAgentURL("ftp://10.0.0.1:8080", true)
	if err == nil {
		t.Error("期望不支持的 scheme 被拦截")
	}
}

func TestValidateAgentURL_EmptyHost(t *testing.T) {
	err := validateAgentURL("http://", false)
	if err == nil {
		t.Error("期望空 host 被拦截")
	}
}

func TestValidateAgentURL_InvalidURL(t *testing.T) {
	err := validateAgentURL("://invalid", false)
	if err == nil {
		t.Error("期望无效 URL 被拦截")
	}
}

func TestValidateAgentURL_DomainResolvesToPublicIP(t *testing.T) {
	// 使用 mock server 验证域名形式地址能被正确处理
	mockAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockAgent.Close()

	// httptest.Server 的 URL 是 127.0.0.1，所以不用 allowPrivate 会失败
	err := validateAgentURL(mockAgent.URL, false)
	if err == nil {
		t.Error("期望 127.0.0.1 被 SSRF 拦截")
	}

	// allowPrivate=true 时应放行
	err = validateAgentURL(mockAgent.URL, true)
	if err != nil {
		t.Errorf("allowPrivate=true 时应放行，但被拦截: %v", err)
	}
}

func TestValidateAgentURL_WSScheme(t *testing.T) {
	// ws:// 和 wss:// 协议应被支持
	err := validateAgentURL("ws://8.8.8.8:8080", false)
	if err != nil {
		t.Errorf("期望 ws:// 协议被支持，但被拦截: %v", err)
	}

	err = validateAgentURL("wss://8.8.8.8:8080", false)
	if err != nil {
		t.Errorf("期望 wss:// 协议被支持，但被拦截: %v", err)
	}
}
