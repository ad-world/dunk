# Dunk TODO

## Product shape

- CLI name: `dunk`.
- Primary UX:
  - `dunk claude`
  - `dunk codex`
  - `dunk pi`
  - `dunk aider`
- MVP user-facing commands:
  - `dunk <software>` — start/reuse the sandbox for the current repo worktree, sync workspace/config, and connect to the sandbox. First pass prints the command to run manually inside the sandbox rather than using shell startup magic.
  - `dunk stop` — kill the active sandbox for the current repo worktree while retaining Dunk's local state so the next run can recreate/reuse as provider capabilities allow.
- Explicitly defer `dunk pull` for now. Do not design or implement pull semantics in the first pass.
- Sandbox identity: one workspace/sandbox per local repo worktree. Multiple software sessions can share that workspace.
- First provider: **E2B sandboxes**, not Fly.io.
- Build in Go with Cobra.
- Use `flake.nix` for development/build environment.
- Initialize git and add README.

## Important architecture correction

Do **not** design this around an `SSHExecutor` or provider-specific executors like `E2BExecutor` as the primary mental model.

Dunk wants a basic provider/runtime abstraction that exposes the operations Dunk needs, regardless of whether the backend is E2B, SSH, Fly, managed infra, or an enterprise adapter.

The abstraction should be about Dunk's required operations:

- ensure a workspace exists
- start/stop it if supported
- upload/sync files
- download/pull files
- run a command
- attach to an interactive session if supported
- report capabilities

Not about the transport implementation.

Transport details such as SSH, E2B SDK calls, websockets, rsync, tar upload, or provider APIs should be private implementation details of each provider/runtime.

## Preferred core interface direction

Sketch only; refine during implementation.

```go
type Runtime interface {
    Name() string

    Ensure(ctx context.Context, req EnsureRequest) (*Workspace, error)
    Stop(ctx context.Context, ws *Workspace, req StopRequest) error

    Push(ctx context.Context, ws *Workspace, manifest TransferManifest) error

    Run(ctx context.Context, ws *Workspace, cmd CommandSpec) (*CommandResult, error)
    Attach(ctx context.Context, ws *Workspace, session SessionSpec, opts AttachOptions) error

    Capabilities(ctx context.Context) RuntimeCapabilities
}
```

Example capability shape:

```go
type RuntimeCapabilities struct {
    CanCreate      bool
    CanStop        bool
    CanPersist     bool
    CanAttachTTY     bool
    CanDetach        bool
    CanReattach      bool
    CanUploadFiles   bool
    MaxLifetime      time.Duration
}
```

Example workspace shape:

```go
type Workspace struct {
    ID       string
    Provider string
    Name     string
    Workdir  string
    ProviderState json.RawMessage // opaque provider-private persisted data; app/CLI must not inspect it
}
```

The E2B implementation can use the E2B SDK internally. A future SSH implementation can use SSH/rsync internally. The CLI/app layer should not care.

## MVP implementation phases

### Phase 0 — scaffold

- [ ] `git init`
- [ ] `go mod init dunk`
- [ ] Add Cobra.
- [ ] Add `gopkg.in/yaml.v3`.
- [ ] Create `flake.nix`.
- [ ] Create `README.md`.
- [ ] Create initial package layout.

Proposed layout:

```text
cmd/dunk/main.go
internal/cli/
internal/app/
internal/config/
internal/project/
internal/runtime/
internal/runtime/e2b/
internal/syncplan/
internal/ftux/
internal/state/
internal/session/
```

### Phase 1 — CLI skeleton

- [x] `dunk <software>` command routing.
- [x] `dunk stop`.
- [x] Helpful `--help` output.
- [x] Wire through app services and E2B runtime.

### Phase 2 — FTUX and config scanner

First run of `dunk claude` should guide the user.

- [ ] Detect missing `dunk.yaml`.
- [ ] Ask to create one.
- [ ] Detect provider credentials:
  - [ ] `E2B_API_KEY`
- [ ] Detect agent env vars:
  - [ ] `ANTHROPIC_API_KEY`
  - [ ] `ANTHROPIC_AUTH_TOKEN`
  - [ ] `OPENAI_API_KEY`
- [ ] Detect project files:
  - [ ] `AGENTS.md`
  - [ ] `CLAUDE.md`
  - [ ] `.mcp.json`
  - [ ] `.claude/**`
  - [ ] `.codex/**`
  - [ ] `.agents/**`
  - [ ] `.pi/**`
- [ ] Detect user config candidates:
  - [ ] `~/.claude/settings.json`
  - [ ] `~/.claude.json`
  - [ ] `~/.codex/config.toml`
  - [ ] `~/.aider.conf.yml`
  - [ ] `~/.gitconfig`
- [ ] Detect credential files but do not copy by default:
  - [ ] `~/.claude/.credentials.json`
  - [ ] `~/.codex/auth.json`
  - [ ] `~/.ssh/id_*`
  - [ ] `.env`
  - [ ] `.env.local`
  - [ ] `~/.npmrc`
- [ ] Detect MCP configs and classify portability issues:
  - [ ] local absolute command paths
  - [ ] missing env vars
  - [ ] local-only services/ports
  - [ ] remote HTTP/SSE servers
- [ ] Make MCP auth/setup easy:
  - [ ] show detected MCP servers during FTUX
  - [ ] show required env vars and missing credentials
  - [ ] help users choose which MCP servers should be enabled remotely
  - [ ] pass selected MCP auth env vars safely to the cloud agent
  - [ ] never silently enable project-controlled MCP commands with copied secrets

### Phase 3 — sync planner

Do not hardcode giant exclude lists in `dunk.yaml`.

Default sync semantics:

- [ ] Respect `.gitignore` by default.
- [ ] Respect `.dunkignore` if present.
- [ ] Let explicit `include` override ignore rules, including `.gitignore` and `.dunkignore` matches.
- [ ] Let explicit `exclude` override includes as the final deny-list.
- [ ] Warn before including likely secrets.
- [ ] Keep generated plan inspectable.
- [ ] Runtime `Push` should receive a resolved `TransferManifest`, not raw `.gitignore`/`.dunkignore` policy.

Sync precedence:

1. enumerate candidate files
2. apply `.gitignore` and `.dunkignore`
3. apply explicit `include` exceptions
4. apply explicit `exclude` as final deny-list
5. run secret/MCP safety warning and approval pass

Possible config shape:

```yaml
sync:
  respect_gitignore: true
  ignore_file: .dunkignore

  include:
    - AGENTS.md
    - CLAUDE.md
    - .mcp.json
    - .claude/**
    - .codex/**
    - .agents/**
    - .pi/**

  exclude: []
```

Need support for users who intentionally want gitignored files loaded into the sandbox; they should just add those paths to `include`. Secret-looking files such as `.env`, `.env.local`, npm tokens, SSH keys, and provider credential files must still require explicit warning/approval before upload.

### Phase 4 — E2B runtime spike

Research and verify before locking implementation:

- [ ] Current E2B Go SDK status/API.
- [ ] Sandbox create/resume semantics.
- [ ] Sandbox persistence and max lifetime.
- [ ] File upload APIs.
- [ ] Whether E2B supports interactive TTY attach.
- [ ] Whether E2B supports detach/reattach to an already-running interactive process.
- [ ] Whether we need to run `tmux` inside the sandbox.
- [ ] Whether custom E2B templates can preinstall `claude`, `codex`, `pi`, etc.
- [ ] Verify E2B's closest operation for the MVP `dunk stop` contract: kill/terminate the active sandbox for this repo worktree while retaining Dunk local state.

### Phase 5 — E2B runtime implementation

- [ ] Implement `internal/runtime/e2b` behind the generic `Runtime` interface.
- [ ] `Ensure` creates/reuses the E2B sandbox for the current repo worktree.
- [ ] `Push` uploads files from a resolved transfer manifest.
- [ ] `Run` executes bootstrap/setup commands.
- [ ] `Attach` currently opens an E2B sandbox shell and prints the command to run. Direct session startup/reattach is deferred until direct E2B PTY support is implemented.
- [ ] `Stop` kills/terminates the active E2B sandbox for this repo worktree while retaining Dunk local state.

### Phase 6 — software profiles

Initial profiles:

- [ ] `claude`
- [ ] `codex`
- [ ] `pi`
- [ ] generic fallback for arbitrary command names

Do not include `aider` in the first iteration.

Example config:

```yaml
software:
  claude:
    command: claude
    env:
      - ANTHROPIC_API_KEY
      - ANTHROPIC_AUTH_TOKEN
  codex:
    command: codex
    env:
      - OPENAI_API_KEY
  pi:
    command: pi
```

### Secret transport rules

- [ ] Never write secret values to `dunk.yaml`.
- [ ] Never write secret values to local state, logs, or inspectable sync plans.
- [ ] Store selected secret **names** in config, not values.
- [ ] Pass selected env var values only to the launched process/session unless the user explicitly opts into persisted remote config.
- [ ] Mask secret values in diagnostics.
- [ ] Ask before uploading any credential-looking file, even if explicitly included.

### Session/attach contract

Before freezing the runtime interface, define transport-neutral session details:

- [ ] `SessionSpec`: session ID/name, command, workdir, env names, TTY requirement.
- [ ] `AttachOptions`: stdin/stdout/stderr ownership, initial terminal size, resize events, detach behavior.
- [ ] capabilities for `CanAttachTTY`, `CanDetach`, and `CanReattach`.
- [ ] direct E2B PTY support should replace the current simple E2B CLI shell bridge.

### Deferred: `dunk pull`

Do not implement `dunk pull` in the first pass. Returning changes from the sandbox to local is a separate design problem.

Notes for later:

- CloudCLI/claudecodeui's `pull` is Git-centric: it shells out to `git pull <remote> <branch>` and surfaces merge-conflict/uncommitted-change errors.
- Dunk's pull problem is different because files are edited in an external sandbox and may include uncommitted, untracked, ignored, or generated files.
- Future options include patch-based pull, provider file download, Git branch/commit workflow, or explicit path download.
- Do not let pull design block the first `dunk <software>` + `dunk stop` MVP.

### Phase 7 — validation

- [ ] `gofmt -w .`
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go run ./cmd/dunk --help`
- [ ] `go run ./cmd/dunk claude --dry-run`
- [ ] `nix flake check`
- [ ] Manual E2B smoke test only when `E2B_API_KEY` is present.

## Non-goals for first iteration

- No Fly.io implementation.
- No SSH provider implementation.
- No managed backend.
- No enterprise external provider protocol execution.
- No web UI.
- No sharing/org/SSO/audit logs.
- No silent credential copying.
- No `dunk pull` implementation in the first pass.
- No broad destructive pull behavior when pull is eventually designed.
- No `aider` first-class profile in the first iteration.

## Open questions

- Does E2B support the exact interactive attach/detach UX needed for `dunk claude`?
- Later: what is the best implementation for conservative `dunk pull`: manifest-based copy, git diff/status inside sandbox, archive download, provider-native file API, or Git branch workflow?
- Should first-run FTUX be fully interactive immediately, or start with `--dry-run` and generated recommendations?
- How much of Claude/Codex/Pi config should be copied vs regenerated remotely?
- Should `dunk.yaml` be committed by default?
