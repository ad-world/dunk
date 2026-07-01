package app

import (
	"path/filepath"

	"dunk/internal/agents"
	"dunk/internal/config"
	"dunk/internal/ftux"
	"dunk/internal/project"
	"dunk/internal/runtime"
	"dunk/internal/state"
	"dunk/internal/syncplan"
)

type RunPlan struct {
	Project      project.Project
	ConfigPath   string
	Config       config.Config
	ConfigExists bool
	Agent        string
	Profile      config.AgentConfig
	Builtin      agents.Profile
	Env          map[string]string
	Findings     []ftux.Finding
	Manifest     runtime.TransferManifest
	Auth         agents.AuthPlan
	SyncWarnings []string
	ProjectState state.ProjectState
}

func BuildRunPlan(agent string, allowSecrets bool) (RunPlan, error) {
	prj, err := project.Detect()
	if err != nil {
		return RunPlan{}, err
	}
	cfgPath := filepath.Join(prj.Root, "dunk.yaml")
	cfg, exists, err := config.Load(cfgPath)
	if err != nil {
		return RunPlan{}, err
	}
	builtin := agents.Builtin(agent)
	profile := cfg.AgentProfile(agent)
	if profile.Command == "" {
		profile.Command = agent
	}
	manifest, warnings, err := syncplan.Build(prj.Root, cfg.Sandbox.Workdir, cfg.Sync, syncplan.Options{AllowSecrets: allowSecrets})
	if err != nil {
		return RunPlan{}, err
	}
	auth, err := agents.BuildAuthPlan(agent)
	if err != nil {
		return RunPlan{}, err
	}
	st, err := state.Load()
	if err != nil {
		return RunPlan{}, err
	}
	return RunPlan{
		Project:      prj,
		ConfigPath:   cfgPath,
		Config:       cfg,
		ConfigExists: exists,
		Agent:        agent,
		Profile:      profile,
		Builtin:      builtin,
		Env:          selectedEnv(profile.Env),
		Findings:     ftux.Scan(prj.Root, profile.Env),
		Manifest:     manifest,
		Auth:         auth,
		SyncWarnings: warnings,
		ProjectState: st.Projects[prj.Key],
	}, nil
}

func (p RunPlan) Session() runtime.SessionSpec {
	return runtime.SessionSpec{Name: "dunk-" + p.Agent, Command: p.Profile.Command, Workdir: p.Config.Sandbox.Workdir, Env: p.Env, TTY: true}
}

func (p RunPlan) EnsureRequest() runtime.EnsureRequest {
	return runtime.EnsureRequest{
		ProjectName: p.Project.Name,
		ProjectRoot: p.Project.Root,
		Workdir:     p.Config.Sandbox.Workdir,
		Template:    p.Config.Sandbox.Template,
		Timeout:     p.Config.TimeoutDuration(),
		Metadata: map[string]string{
			"dunk_project": p.Project.Key,
			"dunk_root":    p.Project.Root,
		},
		State: providerState(p.ProjectState),
	}
}

func providerState(ps state.ProjectState) []byte {
	if ps.Workspace == nil {
		return nil
	}
	return ps.Workspace.ProviderState
}
