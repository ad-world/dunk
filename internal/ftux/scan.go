package ftux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Finding struct {
	Kind, Path, Message string
	Secret              bool
}

func Scan(projectRoot string, envNames []string) []Finding {
	var out []Finding
	for _, e := range append([]string{"E2B_API_KEY"}, envNames...) {
		if os.Getenv(e) != "" {
			out = append(out, Finding{Kind: "env", Path: e, Message: "present"})
		}
	}
	for _, p := range []string{"AGENTS.md", "CLAUDE.md", ".mcp.json", ".claude", ".codex", ".agents", ".pi"} {
		if exists(filepath.Join(projectRoot, p)) {
			out = append(out, Finding{Kind: "project", Path: p, Message: "found"})
		}
	}
	h, _ := os.UserHomeDir()
	for _, p := range []string{".claude/settings.json", ".claude.json", ".codex/config.toml", ".aider.conf.yml", ".gitconfig"} {
		if exists(filepath.Join(h, p)) {
			out = append(out, Finding{Kind: "user-config", Path: "~/" + p, Message: "candidate"})
		}
	}
	for _, p := range []string{".claude/.credentials.json", ".codex/auth.json", ".npmrc"} {
		if exists(filepath.Join(h, p)) {
			out = append(out, Finding{Kind: "credential", Path: "~/" + p, Message: "will not copy by default", Secret: true})
		}
	}
	for _, p := range []string{".env", ".env.local"} {
		if exists(filepath.Join(projectRoot, p)) {
			out = append(out, Finding{Kind: "credential", Path: p, Message: "will not copy by default", Secret: true})
		}
	}
	return out
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func Print(findings []Finding) {
	if len(findings) == 0 {
		fmt.Println("No local agent config/env findings.")
		return
	}
	fmt.Println("Local config scan:")
	for _, f := range findings {
		mark := "✓"
		if f.Secret {
			mark = "⚠"
		}
		fmt.Printf("  %s %-12s %-32s %s\n", mark, f.Kind, f.Path, f.Message)
	}
	fmt.Println(strings.Repeat("-", 60))
}
