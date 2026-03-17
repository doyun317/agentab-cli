#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_DIR="${ROOT_DIR}/.tools"
GO_VERSION="${GO_VERSION:-1.25.8}"

mkdir -p "${TOOLS_DIR}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${OS}" in
  linux) GO_OS="linux" ;;
  darwin) GO_OS="darwin" ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

case "${ARCH}" in
  x86_64|amd64) GO_ARCH="amd64" ;;
  aarch64|arm64) GO_ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

install_go() {
  local url archive target extracted
  url="https://go.dev/dl/go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
  archive="${TOOLS_DIR}/go-${GO_VERSION}.tar.gz"
  target="${TOOLS_DIR}/go"
  extracted="${TOOLS_DIR}/go-extract"

  if [[ -x "${target}/bin/go" ]]; then
    echo "Go already installed at ${target}"
    return
  fi

  echo "Downloading Go ${GO_VERSION} from ${url}"
  curl -fsSL "${url}" -o "${archive}"
  rm -rf "${target}" "${extracted}"
  mkdir -p "${extracted}"
  tar -xzf "${archive}" -C "${extracted}"
  mv "${extracted}/go" "${target}"
  rm -rf "${extracted}"
}

install_go

cat <<EOF
Toolchains installed.
Add to PATH:
  export PATH="${TOOLS_DIR}/go/bin:\$PATH"
EOF
