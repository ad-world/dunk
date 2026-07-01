# dunk

Run terminal coding agents in persistent cloud sandboxes.

```bash
dunk claude
dunk codex
dunk pi
```

MVP scope:

- `dunk <agent>` starts/reuses one E2B sandbox per local repo worktree, uploads selected local files, and opens an E2B sandbox shell. The current first pass prints the command to run inside the sandbox instead of pretending to support real attach/reattach.
- `dunk stop` kills the active sandbox for the current repo worktree while retaining local Dunk state.
- `dunk pull` is intentionally deferred.

## Requirements

- Go 1.22+
- `E2B_API_KEY`
- E2B CLI for interactive attach and command execution:

```bash
brew install e2b
# or
npm i -g @e2b/cli
```

## Development

```bash
nix develop
go run ./cmd/dunk --help
go run ./cmd/dunk claude --dry-run
```

## Config

On first run Dunk can create `dunk.yaml` with defaults. Example:

```yaml
provider: e2b
sandbox:
  template: base
  timeout: 1h
  workdir: /workspace
sync:
  respect_gitignore: true
  include:
    - AGENTS.md
    - CLAUDE.md
    - .mcp.json
    - .claude/**
    - .codex/**
    - .agents/**
    - .pi/**
  exclude: []
agents:
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

Secret values are never written to `dunk.yaml`, local state, or sync plans. Explicitly selected secret-looking files require `--allow-secrets`.
