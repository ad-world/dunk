package agents

import (
	"os"
	"path/filepath"

	"dunk/internal/runtime"
)

type AuthFile struct {
	LocalPath  string
	RemotePath string
	Required   bool
	Sensitive  bool
}

type Profile struct {
	Name      string
	AuthFiles []AuthFile
	Bootstrap []string
}

func Builtin(name string) Profile {
	switch name {
	case "pi":
		return Profile{
			Name: "pi",
			AuthFiles: []AuthFile{
				{LocalPath: "~/.pi/agent/auth.json", RemotePath: "/home/user/.pi/agent/auth.json", Required: true, Sensitive: true},
				{LocalPath: "~/.pi/agent/models.json", RemotePath: "/home/user/.pi/agent/models.json", Required: false, Sensitive: false},
			},
			Bootstrap: []string{ensureNode22Command, installPiCommand},
		}
	default:
		return Profile{Name: name}
	}
}

const ensureNode22Command = `set -eu
node_major() { "$1" -v 2>/dev/null | sed 's/^v//' | cut -d. -f1 || true; }
current_major="$(node_major node)"
if [ -z "$current_major" ] || [ "$current_major" -lt 22 ]; then
  usr_major="$(node_major /usr/bin/node)"
  if [ -z "$usr_major" ] || [ "$usr_major" -lt 22 ]; then
    curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
    sudo apt-get install -y nodejs
  fi
  if [ -x /usr/local/bin/node ] && [ "$(node_major /usr/local/bin/node)" != "" ] && [ "$(node_major /usr/local/bin/node)" -lt 22 ]; then
    sudo mv /usr/local/bin/node /usr/local/bin/node20-old 2>/dev/null || true
    sudo mv /usr/local/bin/npm /usr/local/bin/npm20-old 2>/dev/null || true
    sudo mv /usr/local/bin/npx /usr/local/bin/npx20-old 2>/dev/null || true
  fi
fi
node -v
npm -v`

const installPiCommand = `command -v pi >/dev/null 2>&1 || sudo npm install -g @earendil-works/pi-coding-agent
pi --version || true`

type AuthPlan struct {
	Manifest runtime.TransferManifest
	Copied   []AuthFile
	Missing  []AuthFile
}

func BuildAuthPlan(agent string) (AuthPlan, error) {
	profile := Builtin(agent)
	var plan AuthPlan
	for _, f := range profile.AuthFiles {
		local, err := expandHome(f.LocalPath)
		if err != nil {
			return AuthPlan{}, err
		}
		st, err := os.Stat(local)
		if err != nil || st.IsDir() {
			plan.Missing = append(plan.Missing, f)
			continue
		}
		plan.Copied = append(plan.Copied, f)
		plan.Manifest.Items = append(plan.Manifest.Items, runtime.TransferItem{
			LocalPath:  local,
			RemotePath: f.RemotePath,
			Mode:       "file",
			Size:       st.Size(),
			Sensitive:  f.Sensitive,
		})
	}
	return plan, nil
}

func expandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
