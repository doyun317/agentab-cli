# agentab CLI 설치 및 첫 실행 가이드

상태: 초안  
작성일: 2026-03-17  
마지막 갱신: 2026-03-18 02:33 UTC
목적: `agentab CLI`를 처음 설치하고 실행하는 사용자가 최소한의 브라우저 조작까지 바로 재현할 수 있게 하기 위함

## 1. 무엇이 설치되는가

`agentab`은 다음 구성으로 동작합니다.

- `agentab` CLI 바이너리
- 로컬 daemon
- PinchTab 브라우저 런타임 백엔드
- Chrome 또는 Chromium 계열 브라우저

중요한 점:

- PinchTab은 없으면 `agentab`이 자동 설치를 시도합니다.
- Chrome 또는 Chromium은 머신에 있어야 합니다.
- `CHROME_BIN`이 있으면 그 경로를 우선 사용합니다.

## 2. 빠른 시작

현재 저장소 기준 가장 빠른 시작:

```bash
cd /workspace/agentab-cli
./scripts/bootstrap-tools.sh
export PATH="/workspace/agentab-cli/.tools/go/bin:$PATH"
cd /workspace/agentab-cli
go run ./cmd/agentab doctor
```

빌드된 바이너리로 쓰고 싶다면:

```bash
cd /workspace/agentab-cli
go build -o agentab ./cmd/agentab
./agentab doctor
```

## 3. 실행 전 준비

확인할 항목:

- Go toolchain이 있거나 저장소의 `.tools`를 사용할 수 있다.
- Chrome 또는 Chromium 계열 브라우저가 설치되어 있다.
- 네트워크가 가능하면 PinchTab 자동 설치가 동작한다.

선택 환경 변수:

- `AGENTAB_HOME`
  기본값은 `${HOME}/.agentab`
- `AGENTAB_PINCHTAB_BIN`
  PinchTab 바이너리 경로를 직접 지정
- `CHROME_BIN`
  사용할 Chrome/Chromium 바이너리 경로를 직접 지정
- `PINCHTAB_URL`
  기존 PinchTab 서버를 직접 지정

## 4. 첫 실행 확인

먼저 `doctor`로 현재 상태를 확인합니다.

```bash
agentab --output text doctor
```

또는 저장소에서 직접 실행:

```bash
cd /workspace/agentab-cli
go run ./cmd/agentab doctor
```

정상이라면 아래 항목을 확인할 수 있습니다.

- `agentabHome`
- `logsDir`
- `daemonLogPath`
- `pinchtabLogPath`
- `artifactsDir`
- `managedBinPath`
- `pinchtabURL`
- `pinchtabHealthy`
- `chromeBin`
- `chromeBinFound`
- `chromeBinSource`

예시:

```text
agentab doctor
home: /home/you/.agentab
logs: /home/you/.agentab/logs
artifacts: /home/you/.agentab/artifacts
managed pinchtab bin: /home/you/.agentab/bin/pinchtab
daemon log: /home/you/.agentab/logs/agentab-daemon.log
pinchtab log: /home/you/.agentab/logs/pinchtab.log

chrome
  status: ok
  source: env
  path: /path/to/chrome

pinchtab
  health: ok
  url: http://127.0.0.1:43921

daemon
  status: running
```

## 5. 첫 브라우저 조작

예시 흐름:

```bash
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab text --session demo
agentab tab snapshot --session demo --save
agentab tab screenshot --session demo --save
agentab session stop demo
```

이 흐름에서 기대하는 결과:

- 세션이 생성된다.
- 탭이 열린다.
- 텍스트를 읽을 수 있다.
- snapshot artifact가 `${AGENTAB_HOME}/artifacts/...`에 저장된다.
- screenshot artifact가 `${AGENTAB_HOME}/artifacts/...`에 저장된다.

## 6. artifact 저장 위치

기본 artifact 루트:

- `${AGENTAB_HOME}/artifacts`

예:

- `${AGENTAB_HOME}/artifacts/demo/tab_123/...`

저장 가능한 대표 산출물:

- `agentab tab snapshot --save`
- `agentab tab screenshot --save`
- `agentab tab pdf --save`

명시적 경로에 저장하고 싶다면:

```bash
agentab tab snapshot --session demo --out /tmp/demo-snapshot.json
agentab tab screenshot --session demo --out /tmp/demo.jpg
```

저장 응답 메타데이터:

- 관리형 저장이면 `managed=true`
- `${AGENTAB_HOME}/artifacts` 아래 경로면 `relativePath`가 같이 제공됨
- 생성 시각은 `createdAt`으로 제공됨

## 7. 새 머신에서 권장 확인 순서

릴리스 체크리스트 기준 최소 확인 순서:

```bash
agentab doctor
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab text --session demo
agentab tab screenshot --session demo --save
agentab session stop demo
```

## 8. 로그 위치

대표 로그 파일:

- `${AGENTAB_HOME}/logs/agentab-daemon.log`
- `${AGENTAB_HOME}/logs/pinchtab.log`

문제가 생기면 먼저 이 두 파일과 `agentab doctor` 결과를 같이 확인하는 것이 좋습니다.

## 9. 다음 문서

- CLI 계약 요약: [cli.md](/workspace/agentab-cli/docs/cli.md)
- CLI 스펙 기준선: [cli-spec-sheet.md](/workspace/agentab-cli/docs/cli-spec-sheet.md)
- CLI 트러블슈팅: [cli-troubleshooting.md](/workspace/agentab-cli/docs/cli-troubleshooting.md)
- CLI 릴리스 체크리스트: [cli-release-checklist.md](/workspace/agentab-cli/docs/cli-release-checklist.md)
