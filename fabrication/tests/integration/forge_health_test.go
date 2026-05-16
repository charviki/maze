//go:build integration

package integration

import "testing"

func TestForgeHealth(t *testing.T) {
	health, err := suite.forgeClient.GetHealth()
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("Status = %s, want ok", health.Status)
	}
}
