#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_DIR="${ROOT_DIR}/.tools"

if [[ ! -x "${TOOLS_DIR}/go/bin/go" ]]; then
  "${ROOT_DIR}/scripts/bootstrap-tools.sh"
fi

export PATH="${TOOLS_DIR}/go/bin:${PATH}"

echo "Running Go tests"
(
  cd "${ROOT_DIR}"
  go test ./...
)
