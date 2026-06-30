# Dunk Pi E2B template

This template bakes in the runtime needed for `dunk pi`:

- Node 22
- git/ripgrep/tmux
- `@earendil-works/pi-coding-agent`

Build/publish this template in E2B as `dunk-pi`, then use:

```yaml
sandbox:
  template: dunk-pi
```

The exact E2B template build command depends on the installed E2B CLI version. If the CLI supports Dockerfile-based template builds, run it from this directory and name the resulting template `dunk-pi`.
