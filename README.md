# agentab-cli

`agentab-cli` is the standalone CLI-only project for `agentab`.

The project goal is simple: an agent or a human should be able to drive a real browser through a lightweight local CLI without needing to know PinchTab internals.

## Layout

- `cmd/` implements the CLI entrypoint
- `internal/` implements the daemon, local state, PinchTab installer, and tests
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

## Local Docs Boundary

Operational notes, worklogs, release checklists, and verification documents are kept in a local `docs-local/` directory and are not tracked in the public repository.

Public release history should be read from the GitHub Releases page.

Mode smoke:

```bash
cd /workspace/agentab-cli
AGENTAB_BIN=/workspace/agentab-cli/tmp/release-v0.1.4/extract/agentab \
./scripts/smoke-modes.sh
```

This smoke validates `click`, `type`, `fill`, `press`, and `scroll` in both `headless` and `headed` modes.
