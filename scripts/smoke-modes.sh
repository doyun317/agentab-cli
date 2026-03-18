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

extract_best_ref() {
  local path="$1"
  python3 - "${path}" <<'PY'
import json, sys
obj = json.load(open(sys.argv[1]))
data = obj.get("data") or {}
print(data.get("bestRef") or data.get("best_ref") or "")
PY
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
  local reveal_ref
  reveal_ref="$(extract_best_ref "${prefix}/tab-find.json")"
  run_cli "${home}" tab click --session "smoke-${mode}" --ref "${reveal_ref}" >"${prefix}/tab-click.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-click.json"

  run_cli "${home}" tab find --session "smoke-${mode}" "Type target input" >"${prefix}/tab-find-type.json"
  local type_ref
  type_ref="$(extract_best_ref "${prefix}/tab-find-type.json")"
  run_cli "${home}" tab type --session "smoke-${mode}" "${type_ref}" "typed smoke value" >"${prefix}/tab-type.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-type.json"

  run_cli "${home}" tab fill --session "smoke-${mode}" "#fill-target" "filled smoke value" >"${prefix}/tab-fill.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-fill.json"

  run_cli "${home}" tab find --session "smoke-${mode}" "Press target input" >"${prefix}/tab-find-press.json"
  local press_ref
  press_ref="$(extract_best_ref "${prefix}/tab-find-press.json")"
  run_cli "${home}" tab click --session "smoke-${mode}" --ref "${press_ref}" >"${prefix}/tab-click-press.json"
  run_cli "${home}" tab type --session "smoke-${mode}" "${press_ref}" "enter smoke value" >"${prefix}/tab-type-press.json"
  run_cli "${home}" tab click --session "smoke-${mode}" --ref "${press_ref}" >"${prefix}/tab-click-press-focus.json"
  run_cli "${home}" tab press --session "smoke-${mode}" Enter >"${prefix}/tab-press.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-press.json"

  run_cli "${home}" tab scroll --session "smoke-${mode}" --selector "#scroll-panel" 1600 >"${prefix}/tab-scroll.json"
  run_cli "${home}" tab text --session "smoke-${mode}" >"${prefix}/tab-text-after-scroll.json"
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
    click_text = json.loads((root / mode / "tab-text-after-click.json").read_text())
    type_text = json.loads((root / mode / "tab-text-after-type.json").read_text())
    fill_text = json.loads((root / mode / "tab-text-after-fill.json").read_text())
    press_text = json.loads((root / mode / "tab-text-after-press.json").read_text())
    scroll_result = json.loads((root / mode / "tab-scroll.json").read_text())

    click_ok = "MODE-SMOKE-OK" in json.dumps(click_text)
    type_ok = "TYPE-SMOKE:typed smoke value" in json.dumps(type_text)
    fill_ok = "FILL-SMOKE:filled smoke value" in json.dumps(fill_text)
    press_ok = "PRESS-SMOKE:enter smoke value" in json.dumps(press_text)
    scroll_data = ((scroll_result.get("data") or {}).get("result") or {})
    scroll_ok = scroll_data.get("scrolled") is True and int(scroll_data.get("y") or 0) >= 120

    print(
        f"{mode}: mode={start['data'].get('mode')} "
        f"source={(start.get('diagnostics') or {}).get('source')} "
        f"click_ok={click_ok} type_ok={type_ok} fill_ok={fill_ok} "
        f"press_ok={press_ok} scroll_ok={scroll_ok}"
    )

    if not all((click_ok, type_ok, fill_ok, press_ok, scroll_ok)):
        sys.exit(1)
PY
