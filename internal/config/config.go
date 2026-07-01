package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string                 `yaml:"provider"`
	Sandbox  SandboxConfig          `yaml:"sandbox"`
	Sync     SyncConfig             `yaml:"sync"`
	Agents   map[string]AgentConfig `yaml:"agents"`
}

type SandboxConfig struct {
	Template string `yaml:"template"`
	Timeout  string `yaml:"timeout"`
	Workdir  string `yaml:"workdir"`
}

type SyncConfig struct {
	RespectGitignore bool     `yaml:"respect_gitignore"`
	Include          []string `yaml:"include"`
	Exclude          []string `yaml:"exclude"`
}

type AgentConfig struct {
	Command string   `yaml:"command"`
	Env     []string `yaml:"env"`
}

func Default() Config {
	return Config{
		Provider: "e2b",
		Sandbox:  SandboxConfig{Template: "base", Timeout: "1h", Workdir: "/workspace"},
		Sync: SyncConfig{
			RespectGitignore: true,
			Include: []string{
				"AGENTS.md", "CLAUDE.md", ".mcp.json", ".claude/**", ".codex/**", ".agents/**", ".pi/**",
			},
		},
		Agents: map[string]AgentConfig{
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
		cfg.Sandbox.Timeout = "1h"
	}
	if cfg.Sandbox.Workdir == "" {
		cfg.Sandbox.Workdir = "/workspace"
	}
	if cfg.Agents == nil {
		cfg.Agents = Default().Agents
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
		return time.Hour
	}
	return d
}

func (c Config) AgentProfile(name string) AgentConfig {
	if p, ok := c.Agents[name]; ok {
		return p
	}
	return AgentConfig{Command: name}
}
