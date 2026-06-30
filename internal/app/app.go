package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"dunk/internal/config"
	"dunk/internal/ftux"
	"dunk/internal/project"
	"dunk/internal/runtime"
	e2brt "dunk/internal/runtime/e2b"
	"dunk/internal/state"
)

type App struct {
	Runtime                   runtime.Runtime
	DryRun, Yes, AllowSecrets bool
}

func New(dryRun, yes, allowSecrets bool) *App {
	return &App{Runtime: e2brt.New(), DryRun: dryRun, Yes: yes, AllowSecrets: allowSecrets}
}

func (a *App) RunAgent(ctx context.Context, agent string) error {
	plan, err := BuildRunPlan(agent, a.AllowSecrets)
	if err != nil {
		return err
	}
	if err := a.ensureConfig(plan); err != nil {
		return err
	}
	renderPlan(plan, a.AllowSecrets)
	if a.DryRun {
		fmt.Printf("Dry run: would run %q in session %s using provider %s.\n", plan.Profile.Command, plan.Session().Name, plan.Config.Provider)
		return nil
	}
	if plan.Config.Provider != "e2b" {
		return fmt.Errorf("unsupported provider %q", plan.Config.Provider)
	}
	ws, err := a.Runtime.Ensure(ctx, plan.EnsureRequest())
	if err != nil {
		return err
	}
	if err := saveWorkspace(plan, ws); err != nil {
		return err
	}
	fmt.Println("Sandbox:", ws.ID)
	if err := a.Runtime.Push(ctx, ws, plan.Manifest); err != nil {
		return err
	}
	if len(plan.Auth.Manifest.Items) > 0 {
		if err := a.Runtime.Push(ctx, ws, plan.Auth.Manifest); err != nil {
			return err
		}
	}
	if err := a.bootstrapAgent(ctx, ws, plan); err != nil {
		return err
	}
	return a.Runtime.Attach(ctx, ws, plan.Session(), runtime.AttachOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}

func (a *App) bootstrapAgent(ctx context.Context, ws *runtime.Workspace, plan RunPlan) error {
	for i, cmd := range plan.Builtin.Bootstrap {
		fmt.Printf("Bootstrap %s (%d/%d)...\n", plan.Agent, i+1, len(plan.Builtin.Bootstrap))
		if _, err := a.Runtime.Run(ctx, ws, runtime.CommandSpec{Command: cmd}); err != nil {
			return fmt.Errorf("bootstrap %s: %w", plan.Agent, err)
		}
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	prj, err := project.Detect()
	if err != nil {
		return err
	}
	st, err := state.Load()
	if err != nil {
		return err
	}
	ps := st.Projects[prj.Key]
	if ps.Workspace == nil {
		return fmt.Errorf("no dunk sandbox state for this repo worktree")
	}
	if a.DryRun {
		fmt.Println("Dry run: would kill sandbox", ps.Workspace.ID)
		return nil
	}
	if err := a.Runtime.Stop(ctx, ps.Workspace, runtime.StopRequest{}); err != nil {
		return err
	}
	fmt.Println("Killed sandbox", ps.Workspace.ID, "(local Dunk state retained)")
	return nil
}

func (a *App) ensureConfig(plan RunPlan) error {
	if plan.ConfigExists {
		return nil
	}
	fmt.Println("No dunk.yaml found.")
	if a.DryRun {
		fmt.Println("Would create dunk.yaml with defaults.")
		return nil
	}
	if !a.Yes && !confirm("Create dunk.yaml with defaults? [y/N] ") {
		return fmt.Errorf("dunk.yaml required")
	}
	if err := config.WriteDefault(plan.ConfigPath); err != nil {
		return err
	}
	fmt.Println("Created", plan.ConfigPath)
	return nil
}

func renderPlan(plan RunPlan, allowSecrets bool) {
	ftux.Print(plan.Findings)
	for _, w := range plan.SyncWarnings {
		fmt.Println("⚠", w)
	}
	if len(plan.SyncWarnings) > 0 && !allowSecrets {
		fmt.Println("Secret-looking files were excluded. Use --allow-secrets to include explicitly selected secrets.")
	}
	fmt.Printf("Sync plan: %d files -> %s\n", len(plan.Manifest.Items), plan.Config.Sandbox.Workdir)
	if len(plan.Auth.Copied) > 0 {
		fmt.Printf("Agent auth: %d known auth/config file(s) will be copied to the sandbox.\n", len(plan.Auth.Copied))
	}
	for _, missing := range plan.Auth.Missing {
		if missing.Required {
			fmt.Printf("⚠ required %s auth file not found locally: %s\n", plan.Agent, missing.LocalPath)
		}
	}
}

func saveWorkspace(plan RunPlan, ws *runtime.Workspace) error {
	st, err := state.Load()
	if err != nil {
		return err
	}
	st.Projects[plan.Project.Key] = state.ProjectState{ProjectRoot: plan.Project.Root, Workspace: ws, UpdatedAt: time.Now()}
	if err := state.Save(st); err != nil {
		return fmt.Errorf("save dunk state: %w", err)
	}
	return nil
}

func selectedEnv(names []string) map[string]string {
	m := map[string]string{}
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			m[n] = v
		}
	}
	return m
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	var s string
	_, _ = fmt.Scanln(&s)
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "y" || s == "yes"
}
