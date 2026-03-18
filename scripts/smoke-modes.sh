#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AGENTAB_BIN:-${ROOT_DIR}/agentab}"
SMOKE_ROOT="${AGENTAB_SMOKE_ROOT:-${ROOT_DIR}/tmp/mode-smoke}"
SITE_DIR="${ROOT_DIR}/testdata/smoke"
PATH_BASE="${AGENTAB_SMOKE_PATH:-/usr/bin:/bin}"

find_chrome_bin() {
  if [[ -n "${CHROME_BIN:-}" && -x "${CHROME_BIN}" ]]; then
    printf '%s\n' "${CHROME_BIN}"
    return
  fi
  local candidate
  for candidate in \
    "/root/.agentab/browsers/chrome/linux-146.0.7680.80/chrome-linux64/chrome" \
    "$(command -v google-chrome 2>/dev/null || true)" \
    "$(command -v chromium 2>/dev/null || true)" \
    "$(command -v chromium-browser 2>/dev/null || true)"; do
    if [[ -n "${candidate}" && -x "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return
    fi
  done
  return 1
}

if [[ ! -x "${BIN}" ]]; then
  echo "agentab binary not found: ${BIN}" >&2
  exit 1
fi

CHROME_BIN="$(find_chrome_bin)"
export CHROME_BIN

mkdir -p "${SMOKE_ROOT}"
rm -rf "${SMOKE_ROOT}/headless" "${SMOKE_ROOT}/headed"
mkdir -p "${SMOKE_ROOT}/headless" "${SMOKE_ROOT}/headed"

PORT="$(
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"

python3 -m http.server "${PORT}" --bind 127.0.0.1 --directory "${SITE_DIR}" >"${SMOKE_ROOT}/http.log" 2>&1 &
HTTP_PID=$!

XVFB_PID=""
cleanup() {
  set +e
  if [[ -n "${XVFB_PID}" ]]; then
    kill "${XVFB_PID}" >/dev/null 2>&1 || true
  fi
  kill "${HTTP_PID}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

URL="http://127.0.0.1:${PORT}/"

run_cli() {
  local home="$1"
  shift
  env -u AGENTAB_PINCHTAB_BIN -u PINCHTAB_URL -u PINCHTAB_TOKEN -u AGENTAB_SKIP_INSTALL \
    PATH="${PATH_BASE}" AGENTAB_HOME="${home}" CHROME_BIN="${CHROME_BIN}" \
    "${BIN}" "$@"
}

run_mode() {
  local mode="$1"
  local home="${SMOKE_ROOT}/${mode}"
  local prefix="${SMOKE_ROOT}/${mode}"

  mkdir -p "${home}"
  run_cli "${home}" session start --mode "${mode}" "smoke-${mode}" >"${prefix}/session-start.json"
  run_cli "${home}" tab open --session "smoke-${mode}" "${URL}" >"${prefix}/tab-open.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text.json"
  run_cli "${home}" tab find --session "smoke-${mode}" "Reveal token" >"${prefix}/tab-find.json"
  local ref
  ref="$(
    python3 - "${prefix}/tab-find.json" <<'PY'
import json, sys
obj = json.load(open(sys.argv[1]))
data = obj.get("data") or {}
print(data.get("bestRef") or data.get("best_ref") or "")
PY
  )"
  run_cli "${home}" tab click --session "smoke-${mode}" --ref "${ref}" >"${prefix}/tab-click.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-click.json"
  run_cli "${home}" daemon stop >"${prefix}/daemon-stop.json"
}

run_mode "headless"

if [[ -z "${DISPLAY:-}" ]]; then
  DISPLAY_NUM=99
  while [[ -e "/tmp/.X11-unix/X${DISPLAY_NUM}" ]]; do
    DISPLAY_NUM="$((DISPLAY_NUM + 1))"
  done
  export DISPLAY=":${DISPLAY_NUM}"
  Xvfb "${DISPLAY}" -screen 0 1440x900x24 >"${SMOKE_ROOT}/xvfb.log" 2>&1 &
  XVFB_PID=$!
  sleep 1
  xdpyinfo -display "${DISPLAY}" >/dev/null
fi

run_mode "headed"

python3 - "${SMOKE_ROOT}" <<'PY'
import json
import sys
from pathlib import Path

root = Path(sys.argv[1])
for mode in ("headless", "headed"):
    start = json.loads((root / mode / "session-start.json").read_text())
    text = json.loads((root / mode / "tab-text-after-click.json").read_text())
    payload = json.dumps(text)
    print(f"{mode}: mode={start['data'].get('mode')} source={(start.get('diagnostics') or {}).get('source')} token_ok={'MODE-SMOKE-OK' in payload}")
PY
