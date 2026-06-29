package syncplan

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"dunk/internal/config"
)

func TestIncludeOverridesGitignoreButSecretsRequireOptIn(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".gitignore", ".env.local\n")
	write(t, root, "main.go", "package main\n")
	write(t, root, ".env.local", "TOKEN=secret\n")
	run(t, root, "git", "init")
	run(t, root, "git", "add", ".gitignore", "main.go")

	cfg := config.SyncConfig{RespectGitignore: true, Include: []string{".env.local"}}
	manifest, warnings, err := Build(root, "/workspace", cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected secret warning")
	}
	for _, item := range manifest.Items {
		if filepath.Base(item.LocalPath) == ".env.local" {
			t.Fatalf("secret file included without opt-in")
		}
	}

	manifest, _, err = Build(root, "/workspace", cfg, Options{AllowSecrets: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range manifest.Items {
		if filepath.Base(item.LocalPath) == ".env.local" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected explicitly included secret with opt-in")
	}
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
