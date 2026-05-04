package casbin

import (
	"testing"
)

func TestEnforce_AdminWildcard(t *testing.T) {
	e, err := NewEnforcer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := SeedPolicies(e); err != nil {
		t.Fatal(err)
	}

	ok, err := e.Enforce(AdminSubjectKey, "host/demo", "read")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("admin should have wildcard access")
	}
}

func TestEnforce_SubjectPolicy(t *testing.T) {
	e, err := NewEnforcer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.AddPolicy("host:demo", "session/*", "connect"); err != nil {
		t.Fatal(err)
	}

	ok, err := e.Enforce("host:demo", "session/abc", "connect")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("host should match subject policy")
	}

	ok, err = e.Enforce("host:demo", "session/abc", "write")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("host should not match wrong action")
	}
}
