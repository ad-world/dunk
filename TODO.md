# Dunk TODO

## Current MVP behavior

- `dunk <agent>` prepares/reuses one E2B sandbox per local repo worktree, uploads the selected workspace files, opens an E2B sandbox shell, and prints the command to run inside the sandbox.
- `dunk stop` kills the active E2B sandbox for the current repo worktree while retaining local Dunk state.
- `dunk pull` is intentionally deferred.

## Remaining near-term work

- Replace the E2B CLI shell bridge with direct E2B PTY support in Go.
- Implement real software session attach/reattach for `dunk claude`, `dunk codex`, and `dunk pi`.
- Build a real FTUX flow for selecting provider, sync policy, env vars, and MCP servers.
- Parse Claude/Codex/Pi MCP config and surface auth/env/path portability issues.
- Decide whether Dunk needs a separate ignore file later; current config uses `.gitignore` plus explicit `include`/`exclude` only.
- Design `dunk pull` separately: patch, branch/commit workflow, explicit file download, or provider file API.
- Run a live E2B smoke test with `E2B_API_KEY` and E2B CLI installed.

## Non-goals for this pass

- No Fly implementation.
- No SSH provider.
- No managed backend.
- No web UI.
- No silent credential copying.
- No broad destructive pull.
