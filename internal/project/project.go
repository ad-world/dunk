package project

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Project struct{ Name, Root, Key string }

func Detect() (Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Project{}, err
	}
	root := cwd
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	if out, err := cmd.Output(); err == nil {
		root = strings.TrimSpace(string(out))
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return Project{}, err
	}
	name := filepath.Base(abs)
	h := sha256.Sum256([]byte(abs))
	return Project{Name: name, Root: abs, Key: name + "-" + hex.EncodeToString(h[:])[:10]}, nil
}
