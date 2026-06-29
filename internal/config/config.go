package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string                    `yaml:"provider"`
	Sandbox  SandboxConfig             `yaml:"sandbox"`
	Sync     SyncConfig                `yaml:"sync"`
	Software map[string]SoftwareConfig `yaml:"software"`
}

type SandboxConfig struct {
	Template string `yaml:"template"`
	Timeout  string `yaml:"timeout"`
	Workdir  string `yaml:"workdir"`
}

type SyncConfig struct {
	RespectGitignore bool     `yaml:"respect_gitignore"`
	IgnoreFile       string   `yaml:"ignore_file"`
	Include          []string `yaml:"include"`
	Exclude          []string `yaml:"exclude"`
}

type SoftwareConfig struct {
	Command string   `yaml:"command"`
	Env     []string `yaml:"env"`
}

func Default() Config {
	return Config{
		Provider: "e2b",
		Sandbox:  SandboxConfig{Template: "base", Timeout: "8h", Workdir: "/workspace"},
		Sync: SyncConfig{
			RespectGitignore: true,
			IgnoreFile:       ".dunkignore",
			Include: []string{
				"AGENTS.md", "CLAUDE.md", ".mcp.json", ".claude/**", ".codex/**", ".agents/**", ".pi/**",
			},
		},
		Software: map[string]SoftwareConfig{
			"claude": {Command: "claude", Env: []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"}},
			"codex":  {Command: "codex", Env: []string{"OPENAI_API_KEY"}},
			"pi":     {Command: "pi"},
		},
	}
}

func Load(path string) (Config, bool, error) {
	cfg := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, false, nil
		}
		return cfg, false, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, true, err
	}
	if cfg.Provider == "" {
		cfg.Provider = "e2b"
	}
	if cfg.Sandbox.Template == "" {
		cfg.Sandbox.Template = "base"
	}
	if cfg.Sandbox.Timeout == "" {
		cfg.Sandbox.Timeout = "8h"
	}
	if cfg.Sandbox.Workdir == "" {
		cfg.Sandbox.Workdir = "/workspace"
	}
	if cfg.Software == nil {
		cfg.Software = Default().Software
	}
	return cfg, true, nil
}

func WriteDefault(path string) error {
	cfg := Default()
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func (c Config) TimeoutDuration() time.Duration {
	d, err := time.ParseDuration(c.Sandbox.Timeout)
	if err != nil {
		return 8 * time.Hour
	}
	return d
}

func (c Config) SoftwareProfile(name string) SoftwareConfig {
	if p, ok := c.Software[name]; ok {
		return p
	}
	return SoftwareConfig{Command: name}
}
