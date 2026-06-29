package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"dunk/internal/runtime"
)

type File struct {
	Projects map[string]ProjectState `json:"projects"`
}
type ProjectState struct {
	ProjectRoot string             `json:"project_root"`
	Workspace   *runtime.Workspace `json:"workspace,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func Path() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".dunk", "state.json"), nil
}

func Load() (File, error) {
	p, err := Path()
	if err != nil {
		return File{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return File{Projects: map[string]ProjectState{}}, nil
		}
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return File{}, err
	}
	if f.Projects == nil {
		f.Projects = map[string]ProjectState{}
	}
	return f, nil
}

func Save(f File) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}
