#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_DIR="${ROOT_DIR}/.tools"
CONFIG_PATH="${ROOT_DIR}/.goreleaser.yaml"

"${ROOT_DIR}/scripts/install-goreleaser.sh"

export PATH="${TOOLS_DIR}/go/bin:${TOOLS_DIR}/bin:${PATH}"

cd "${ROOT_DIR}"

if git rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "Validating GoReleaser configuration"
  goreleaser check --config "${CONFIG_PATH}"
else
  echo "No git repository detected: skipping 'goreleaser check' and running snapshot release directly"
fi

goreleaser release --snapshot --clean --config "${CONFIG_PATH}"
cat <<EOF
Snapshot release artifacts available under ${ROOT_DIR}/dist

Notes:
- Snapshot release creates clean archive names and checksums even before a GitHub repo exists.
- Internal build directories under dist/ may still include target suffixes like amd64_v1 or arm64_v8.0.
- Once git is initialized, rerun this script to enable 'goreleaser check' before the same snapshot release flow.
EOF
