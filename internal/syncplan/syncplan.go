package syncplan

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"dunk/internal/config"
	"dunk/internal/runtime"
)

type Options struct{ AllowSecrets bool }

func Build(root, workdir string, cfg config.SyncConfig, opts Options) (runtime.TransferManifest, []string, error) {
	files := map[string]bool{}
	if cfg.RespectGitignore && isGitRepo(root) {
		cmd := exec.Command("git", "ls-files", "-co", "--exclude-standard")
		cmd.Dir = root
		out, err := cmd.Output()
		if err != nil {
			return runtime.TransferManifest{}, nil, err
		}
		s := bufio.NewScanner(strings.NewReader(string(out)))
		for s.Scan() {
			if s.Text() != "" {
				files[filepath.ToSlash(s.Text())] = true
			}
		}
	} else {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || path == root {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			if d.IsDir() && (rel == ".git" || rel == "node_modules") {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				files[rel] = true
			}
			return nil
		})
	}
	if cfg.IgnoreFile != "" {
		applyDunkIgnore(root, cfg.IgnoreFile, files)
	}
	for _, pat := range cfg.Include {
		addMatches(root, pat, files)
	}
	for _, pat := range cfg.Exclude {
		removeMatches(pat, files)
	}
	var rels []string
	for rel := range files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	var warnings []string
	var items []runtime.TransferItem
	for _, rel := range rels {
		if looksSecret(rel) {
			warnings = append(warnings, fmt.Sprintf("secret-looking file selected: %s", rel))
			if !opts.AllowSecrets {
				continue
			}
		}
		st, err := os.Stat(filepath.Join(root, rel))
		if err != nil || st.IsDir() {
			continue
		}
		items = append(items, runtime.TransferItem{LocalPath: filepath.Join(root, rel), RemotePath: filepath.ToSlash(filepath.Join(workdir, rel)), Mode: "file", Size: st.Size()})
	}
	return runtime.TransferManifest{Items: items}, warnings, nil
}

func isGitRepo(root string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = root
	return cmd.Run() == nil
}

func addMatches(root, pattern string, files map[string]bool) {
	pattern = filepath.ToSlash(pattern)
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		filepath.WalkDir(filepath.Join(root, filepath.FromSlash(prefix)), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			files[filepath.ToSlash(rel)] = true
			return nil
		})
		return
	}
	matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
	for _, m := range matches {
		if st, err := os.Stat(m); err == nil && !st.IsDir() {
			rel, _ := filepath.Rel(root, m)
			files[filepath.ToSlash(rel)] = true
		}
	}
}

func removeMatches(pattern string, files map[string]bool) {
	pattern = filepath.ToSlash(pattern)
	for rel := range files {
		if match(pattern, rel) {
			delete(files, rel)
		}
	}
}

func match(pattern, rel string) bool {
	if strings.HasSuffix(pattern, "/**") {
		return strings.HasPrefix(rel, strings.TrimSuffix(pattern, "/**")+"/")
	}
	ok, _ := filepath.Match(pattern, rel)
	return ok || rel == pattern
}

func applyDunkIgnore(root, name string, files map[string]bool) {
	b, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return
	}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		removeMatches(line, files)
	}
}

func looksSecret(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	rel = strings.ToLower(rel)
	if base == ".env" || strings.HasPrefix(base, ".env.") || strings.Contains(rel, "credentials") || strings.Contains(rel, "auth.json") || strings.Contains(rel, "id_rsa") || strings.Contains(rel, "id_ed25519") || base == ".npmrc" {
		return true
	}
	return false
}
