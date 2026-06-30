package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPiAuthPlanCopiesKnownFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	write(t, home, ".pi/agent/auth.json", "{}")
	write(t, home, ".pi/agent/models.json", "{}")

	plan, err := BuildAuthPlan("pi")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Missing) != 0 {
		t.Fatalf("missing = %#v, want none", plan.Missing)
	}
	if len(plan.Manifest.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(plan.Manifest.Items))
	}
	if !plan.Manifest.Items[0].Sensitive {
		t.Fatalf("auth.json should be sensitive")
	}
}

func TestBuildPiAuthPlanReportsMissingRequiredAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	plan, err := BuildAuthPlan("pi")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Missing) == 0 || !plan.Missing[0].Required {
		t.Fatalf("expected missing required auth, got %#v", plan.Missing)
	}
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
