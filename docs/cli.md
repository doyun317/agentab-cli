# agentab CLI

`agentab` is JSON-first. Every command returns the same envelope:

```json
{
  "ok": true,
  "data": {},
  "error": null,
  "diagnostics": {}
}
```

Core command groups:

- `agentab doctor`
- `agentab daemon start|status|stop`
- `agentab session start|list|resume|stop`
- `agentab tab open|list|close|focus|snapshot|text|find|click|type|fill|press|hover|scroll|select|eval|screenshot|pdf`

Examples:

```bash
agentab --output text doctor
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab snapshot --session demo
agentab tab snapshot --session demo --save
agentab tab find --session demo "More information"
agentab tab screenshot --session demo --save
agentab tab click --session demo --tab tab_123 --ref e5
```

Global flags:

- `--session`
- `--tab`
- `--profile`
- `--mode`
- `--owner`
- `--timeout`
- `--output`
- `--debug`

Text output notes:

- `agentab --output text doctor` prints a human-friendly health summary instead of raw JSON.
- `--output json` keeps the full machine-readable envelope for scripts and agents.

Artifact options:

- `agentab tab snapshot --save` writes a managed artifact under `${AGENTAB_HOME}/artifacts/...`
- `agentab tab snapshot --out /path/to/file.json` writes to an explicit path
- `agentab tab screenshot --save` writes a managed JPEG artifact
- `agentab tab pdf --save` writes a managed PDF artifact

Further guides:

- install and first run: [cli-install.md](/workspace/agentab-cli/docs/cli-install.md)
- troubleshooting: [cli-troubleshooting.md](/workspace/agentab-cli/docs/cli-troubleshooting.md)
- release gate: [cli-release-checklist.md](/workspace/agentab-cli/docs/cli-release-checklist.md)
