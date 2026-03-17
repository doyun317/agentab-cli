#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_DIR="${ROOT_DIR}/.tools"
GORELEASER_VERSION="${GORELEASER_VERSION:-v2.12.7}"

if [[ ! -x "${TOOLS_DIR}/go/bin/go" ]]; then
  "${ROOT_DIR}/scripts/bootstrap-tools.sh"
fi

mkdir -p "${TOOLS_DIR}/bin"

export PATH="${TOOLS_DIR}/go/bin:${PATH}"
export GOBIN="${TOOLS_DIR}/bin"

TARGET="${GOBIN}/goreleaser"
if [[ -x "${TARGET}" ]]; then
  INSTALLED="$("${TARGET}" --version 2>/dev/null | awk '/GitVersion:/ { print $2 }')"
  if [[ "${INSTALLED}" == "${GORELEASER_VERSION}" ]]; then
    echo "GoReleaser already installed at ${TARGET} (${INSTALLED})"
    exit 0
  fi
fi

echo "Installing GoReleaser ${GORELEASER_VERSION}"
go install "github.com/goreleaser/goreleaser/v2@${GORELEASER_VERSION}"
echo "Installed GoReleaser at ${TARGET}"
