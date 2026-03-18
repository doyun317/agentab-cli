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

## Official CLI Examples

The GitHub-facing canonical examples live in this README.

Check the local runtime:

```bash
agentab --output text doctor
```

Start a session and open a page:

```bash
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab text --session demo
```

Find and click something:

```bash
agentab tab find --session demo "More information"
agentab tab click --session demo --tab <tab-id> --ref <ref>
agentab tab text --session demo
```

Save artifacts:

```bash
agentab tab snapshot --session demo --save
agentab tab screenshot --session demo --save
```

Clean up:

```bash
agentab session stop demo
agentab daemon stop
```

## More Docs

Deeper local docs:

- [CLI overview](docs/cli.md)
- [CLI install and first run](docs/cli-install.md)
- [CLI troubleshooting](docs/cli-troubleshooting.md)
- [CLI operations runbook](docs/cli-operations-runbook.md)
- [CLI release workflow](docs/cli-release-workflow.md)
- [CLI release checklist](docs/cli-release-checklist.md)
- [Release verification history](docs/releases/README.md)

Mode smoke:

```bash
cd /workspace/agentab-cli
AGENTAB_BIN=/workspace/agentab-cli/tmp/release-v0.1.4/extract/agentab \
./scripts/smoke-modes.sh
```

This smoke validates `click`, `type`, `fill`, `press`, and `scroll` in both `headless` and `headed` modes.
