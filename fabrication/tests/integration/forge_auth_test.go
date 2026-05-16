//go:build integration

package integration

import "testing"

func TestForgeAuthRequired(t *testing.T) {
	// 不带 Authorization 的请求应返回 401
	status, err := suite.forgeClient.GetUnauthenticated("/api/v1/archives")
	if err != nil {
		t.Fatalf("unauthenticated request: %v", err)
	}
	if status != 401 {
		t.Errorf("unauthenticated status = %d, want 401", status)
	}

	// 带 JWT 的请求应正常
	_, err = suite.forgeClient.ListArchives()
	if err != nil {
		t.Fatalf("authenticated ListArchives: %v", err)
	}
}
