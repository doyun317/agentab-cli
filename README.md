# agentab-cli

`agentab-cli` is the standalone CLI-only project for `agentab`.

The project goal is simple: an agent or a human should be able to drive a real browser through a lightweight local CLI without needing to know PinchTab internals.

## Layout

- `cmd/` implements the CLI entrypoint
- `internal/` implements the daemon, local state, PinchTab installer, and tests
- `docs/` contains CLI guides and release notes
- `scripts/` contains local toolchain bootstrap and test helpers

## Quick Start

```bash
cd /workspace/agentab-cli
./scripts/bootstrap-tools.sh
./scripts/test.sh
```

With the local toolchains installed, you can run:

```bash
export PATH="/workspace/agentab-cli/.tools/go/bin:$PATH"
cd /workspace/agentab-cli
go run ./cmd/agentab --output text doctor
```

CLI-focused docs:

- [CLI overview](/workspace/agentab-cli/docs/cli.md)
- [CLI install and first run](/workspace/agentab-cli/docs/cli-install.md)
- [CLI troubleshooting](/workspace/agentab-cli/docs/cli-troubleshooting.md)
- [CLI operations runbook](/workspace/agentab-cli/docs/cli-operations-runbook.md)
- [CLI release workflow](/workspace/agentab-cli/docs/cli-release-workflow.md)
- [CLI release checklist](/workspace/agentab-cli/docs/cli-release-checklist.md)
- [Release verification history](/workspace/agentab-cli/docs/releases/README.md)

Mode smoke:

```bash
cd /workspace/agentab-cli
AGENTAB_BIN=/workspace/agentab-cli/tmp/release-v0.1.4/extract/agentab \
./scripts/smoke-modes.sh
```

This smoke validates `click`, `type`, `fill`, `press`, and `scroll` in both `headless` and `headed` modes.
