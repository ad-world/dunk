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

const MaxUploadSize int64 = 50 << 20

type Options struct{ AllowSecrets bool }

type Warning string

func Build(root, workdir string, cfg config.SyncConfig, opts Options) (runtime.TransferManifest, []string, error) {
	files, err := discover(root, cfg)
	if err != nil {
		return runtime.TransferManifest{}, nil, err
	}
	applyIncludes(root, cfg.Include, files)
	applyExcludes(cfg.Exclude, files)
	return buildManifest(root, workdir, files, opts)
}

func discover(root string, cfg config.SyncConfig) (map[string]bool, error) {
	if cfg.RespectGitignore && isGitRepo(root) {
		return discoverGit(root)
	}
	return discoverFilesystem(root)
}

func discoverGit(root string) (map[string]bool, error) {
	cmd := exec.Command("git", "ls-files", "-co", "--exclude-standard")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := map[string]bool{}
	s := bufio.NewScanner(strings.NewReader(string(out)))
	for s.Scan() {
		if s.Text() != "" {
			files[filepath.ToSlash(s.Text())] = true
		}
	}
	return files, s.Err()
}

func discoverFilesystem(root string) (map[string]bool, error) {
	files := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == root {
			return err
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
	return files, err
}

func applyIncludes(root string, patterns []string, files map[string]bool) {
	for _, pat := range patterns {
		addMatches(root, pat, files)
	}
}

func applyExcludes(patterns []string, files map[string]bool) {
	for _, pat := range patterns {
		removeMatches(pat, files)
	}
}

func buildManifest(root, workdir string, files map[string]bool, opts Options) (runtime.TransferManifest, []string, error) {
	rels := sorted(files)
	var warnings []string
	var items []runtime.TransferItem
	for _, rel := range rels {
		local := filepath.Join(root, rel)
		st, err := os.Stat(local)
		if err != nil || st.IsDir() {
			continue
		}
		if looksSecret(rel) {
			warnings = append(warnings, fmt.Sprintf("secret-looking file selected: %s", rel))
			if !opts.AllowSecrets {
				continue
			}
		}
		if st.Size() > MaxUploadSize {
			warnings = append(warnings, fmt.Sprintf("file exceeds %d MiB upload limit and was skipped: %s", MaxUploadSize>>20, rel))
			continue
		}
		items = append(items, runtime.TransferItem{LocalPath: local, RemotePath: filepath.ToSlash(filepath.Join(workdir, rel)), Mode: "file", Size: st.Size()})
	}
	return runtime.TransferManifest{Items: items}, warnings, nil
}

func sorted(files map[string]bool) []string {
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	return rels
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

func looksSecret(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	rel = strings.ToLower(rel)
	return base == ".env" || (strings.HasPrefix(base, ".env.") && base != ".env.example") || strings.Contains(rel, "credentials") || strings.Contains(rel, "auth.json") || strings.Contains(rel, "id_rsa") || strings.Contains(rel, "id_ed25519") || base == ".npmrc"
}
