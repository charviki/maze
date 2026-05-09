//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/auth"
	"github.com/charviki/maze-integration-tests/kit"
)

func authLoginURL() string {
	return suite.cfg.DirectorCoreURL + "/api/v1/auth/login"
}

func authRefreshURL() string {
	return suite.cfg.DirectorCoreURL + "/api/v1/auth/refresh"
}

func authLogoutURL() string {
	return suite.cfg.DirectorCoreURL + "/api/v1/auth/logout"
}

func doJSONPostWithHeaders(t *testing.T, url string, payload interface{}, headers map[string]string) (*http.Response, []byte, map[string]interface{}) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	var result map[string]interface{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
	return resp, raw, result
}

func doJSONPost(t *testing.T, url string, payload interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()
	resp, _, result := doJSONPostWithHeaders(t, url, payload, nil)
	return resp, result
}

func authHeader(accessToken string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + accessToken,
	}
}

func assertAuthJSONUsesCamelCase(t *testing.T, result map[string]interface{}) {
	t.Helper()
	if _, ok := result["accessToken"]; !ok {
		t.Fatal("expected camelCase field accessToken in JSON response")
	}
	if _, ok := result["refreshToken"]; !ok {
		t.Fatal("expected camelCase field refreshToken in JSON response")
	}
	if _, ok := result["expiresIn"]; !ok {
		t.Fatal("expected camelCase field expiresIn in JSON response")
	}
	if _, ok := result["access_token"]; ok {
		t.Fatal("unexpected snake_case field access_token in JSON response")
	}
	if _, ok := result["refresh_token"]; ok {
		t.Fatal("unexpected snake_case field refresh_token in JSON response")
	}
	if _, ok := result["expires_in"]; ok {
		t.Fatal("unexpected snake_case field expires_in in JSON response")
	}
}

func TestAuth_LoginSuccess(t *testing.T) {
	resp, result := doJSONPost(t, authLoginURL(), map[string]string{
		"username": suite.cfg.AdminUsername,
		"password": suite.cfg.AdminPassword,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}
	assertAuthJSONUsesCamelCase(t, result)

	accessToken, _ := result["accessToken"].(string)
	refreshToken, _ := result["refreshToken"].(string)
	expiresInRaw, _ := result["expiresIn"].(string)
	expiresIn, err := strconv.ParseInt(expiresInRaw, 10, 64)
	if err != nil {
		t.Fatalf("expected expiresIn to be a numeric string, got %q: %v", expiresInRaw, err)
	}

	if accessToken == "" {
		t.Error("expected non-empty accessToken")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refreshToken")
	}
	if expiresIn <= 0 {
		t.Error("expected positive expiresIn")
	}
	t.Logf("[step] PASS: login returned accessToken (len=%d) refreshToken (len=%d) expiresIn=%d",
		len(accessToken), len(refreshToken), expiresIn)
}

func TestAuth_LoginInvalidCredentials(t *testing.T) {
	resp, _ := doJSONPost(t, authLoginURL(), map[string]string{
		"username": suite.cfg.AdminUsername,
		"password": "wrong-password",
	})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	t.Log("[step] PASS: invalid credentials rejected with 401")
}

func TestAuth_LoginMissingFields(t *testing.T) {
	resp, _ := doJSONPost(t, authLoginURL(), map[string]string{})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	t.Log("[step] PASS: missing fields rejected with 400")
}

func TestAuth_ProtectedEndpointWithoutToken(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, suite.cfg.DirectorCoreURL+"/api/v1/hosts", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated request, got %d", resp.StatusCode)
	}
	t.Log("[step] PASS: protected endpoint rejects request without token")
}

func TestAuth_TokenRefresh(t *testing.T) {
	loginResult, err := kit.LoginAdmin(context.Background(), suite.cfg)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, result := doJSONPost(t, authRefreshURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %v", resp.StatusCode, result)
	}

	newAccessToken, _ := result["accessToken"].(string)
	newRefreshToken, _ := result["refreshToken"].(string)

	if newAccessToken == "" {
		t.Error("expected non-empty new accessToken")
	}
	if newRefreshToken == "" {
		t.Error("expected non-empty new refreshToken")
	}
	if newAccessToken == loginResult.AccessToken {
		t.Error("refresh should issue a new access token")
	}
	t.Log("[step] PASS: token refresh issued new token pair")
}

func TestAuth_RefreshWithRevokedToken(t *testing.T) {
	loginResult, err := kit.LoginAdmin(context.Background(), suite.cfg)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, _ := doJSONPost(t, authRefreshURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first refresh failed: got %d", resp.StatusCode)
	}

	resp, _ = doJSONPost(t, authRefreshURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for replayed refresh token, got %d", resp.StatusCode)
	}
	t.Log("[step] PASS: replayed refresh token correctly rejected")
}

func TestAuth_LogoutAcceptsCamelCaseRefreshToken(t *testing.T) {
	loginResult, err := kit.LoginAdmin(context.Background(), suite.cfg)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, raw, result := doJSONPostWithHeaders(t, authLogoutURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	}, authHeader(loginResult.AccessToken))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(raw))
	}
	if len(result) != 0 {
		t.Fatalf("expected empty logout response body, got %v", result)
	}

	resp, _ = doJSONPost(t, authRefreshURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout revoked refresh token, got %d", resp.StatusCode)
	}
	t.Log("[step] PASS: logout accepts camelCase refreshToken and revokes the caller's refresh token")
}

func TestAuth_LogoutWithForeignTokenDoesNotRevokeOwnedRefreshToken(t *testing.T) {
	loginResult, err := kit.LoginAdmin(context.Background(), suite.cfg)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	if suite.cfg.JWTSecret == "" {
		t.Fatal("MAZE_TEST_JWT_SECRET is required for foreign-token logout verification")
	}

	// 这里显式伪造一个“其他主体”的访问令牌：
	// Logout 只依赖 JWT 中的 subject，且跳过资源授权映射，因此这是最小且稳定的黑盒集成构造方式。
	foreignAccessToken, err := auth.GenerateAccessToken(
		suite.cfg.JWTSecret,
		auth.DefaultIssuer,
		"user:foreign-logout-"+time.Now().UTC().Format("20060102150405"),
		time.Minute,
	)
	if err != nil {
		t.Fatalf("generate foreign access token: %v", err)
	}

	resp, raw, _ := doJSONPostWithHeaders(t, authLogoutURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	}, authHeader(foreignAccessToken))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for foreign-token logout, got %d: %s", resp.StatusCode, string(raw))
	}

	resp, result := doJSONPost(t, authRefreshURL(), map[string]string{
		"refreshToken": loginResult.RefreshToken,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 because foreign logout must not revoke token, got %d: %v", resp.StatusCode, result)
	}
	if refreshedToken, _ := result["refreshToken"].(string); refreshedToken == "" {
		t.Fatal("expected refresh to succeed after foreign-token logout attempt")
	}
	t.Log("[step] PASS: foreign subject cannot revoke another subject's refresh token via logout")
}
