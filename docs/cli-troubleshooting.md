# agentab CLI 트러블슈팅

상태: 초안  
작성일: 2026-03-17  
마지막 갱신: 2026-03-18 02:33 UTC
목적: `agentab CLI` 사용 중 자주 나오는 설치, daemon, 브라우저, 상태 관련 문제를 빠르게 해결하기 위함

## 1. 먼저 확인할 것

문제가 생기면 가장 먼저 아래 3가지를 확인합니다.

1. `agentab --output text doctor`
2. `${AGENTAB_HOME}/logs/agentab-daemon.log`
3. `${AGENTAB_HOME}/logs/pinchtab.log`

추가로 artifact 경로도 확인할 수 있습니다.

- `${AGENTAB_HOME}/artifacts`

## 2. 자주 나오는 증상과 해결

### 2.1 `no current session; pass --session`

뜻:

- 현재 세션이 없는데 `--session` 없이 명령을 실행한 상태입니다.

해결:

```bash
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab text --session demo
```

또는 항상 `--session <name>`을 명시합니다.

### 2.2 `no current tab; pass --tab or open/focus a tab first`

뜻:

- 현재 탭이 선택되지 않았거나 아직 탭을 열지 않았습니다.

해결:

```bash
agentab tab open --session demo https://example.com
agentab tab list --session demo
agentab tab focus --session demo --tab <tabId>
```

### 2.3 `pinchtab not found and installation skipped`

뜻:

- PinchTab이 없고 자동 설치도 건너뛰도록 설정된 상태입니다.

확인:

- `AGENTAB_SKIP_INSTALL=1`이 설정되어 있는지 확인

해결:

- `AGENTAB_SKIP_INSTALL`을 제거하고 다시 실행
- 또는 `AGENTAB_PINCHTAB_BIN`으로 PinchTab 경로를 직접 지정

### 2.4 `agentab daemon did not start; see .../agentab-daemon.log`

뜻:

- CLI가 daemon auto-start를 시도했지만 지정 시간 안에 health check가 통과하지 않았습니다.

확인:

- `${AGENTAB_HOME}/logs/agentab-daemon.log`
- `${AGENTAB_HOME}/run/daemon.json`

해결:

- 기존 daemon 상태가 꼬였으면 `agentab daemon stop` 실행
- 안 되면 `${AGENTAB_HOME}/run/daemon.json` 상태를 확인
- 포트 충돌, 권한 문제, 손상된 상태 파일 여부 확인

### 2.5 `pinchtab did not become healthy; see .../pinchtab.log`

뜻:

- PinchTab 프로세스는 시작했지만 health check가 정상화되지 않았습니다.

확인:

- `${AGENTAB_HOME}/logs/pinchtab.log`
- `CHROME_BIN` 경로
- `PINCHTAB_URL`이 올바른지

해결:

- 브라우저 바이너리 경로를 `CHROME_BIN`으로 명시
- 로컬 PinchTab이 아니라 원격 URL을 써야 한다면 `PINCHTAB_URL` 확인
- 오래된 daemon이나 PinchTab 프로세스가 남아 있으면 정리 후 재실행

### 2.6 `chromeBin`이 비어 있거나 브라우저가 안 뜬다

뜻:

- `doctor` 기준으로 Chrome/Chromium 경로를 찾지 못했거나, PinchTab이 브라우저 실행에 실패한 상태입니다.
- `CHROME_BIN`을 지정했다면 `chromeBinSource`가 `env`로, PATH에서 찾았다면 `path`로 보입니다.

해결:

- `CHROME_BIN`으로 명시적 경로를 지정
- 머신에 Chrome 또는 Chromium 설치 확인
- Linux라면 shared library 부족 여부 확인

예:

```bash
export CHROME_BIN=/path/to/chrome
agentab doctor
```

확인 포인트:

- `chromeBin`이 기대한 경로인지
- `chromeBinFound`가 `true`인지
- `chromeBinSource`가 `env` 또는 `path` 중 기대한 값인지

### 2.7 Linux에서 `libglib-2.0.so.0` 같은 오류

뜻:

- Chrome 실행에 필요한 시스템 라이브러리가 부족합니다.

해결:

- 브라우저 런타임 패키지를 설치
- 필요한 shared library를 OS 패키지 매니저로 보강

확인:

- `${AGENTAB_HOME}/logs/pinchtab.log`

### 2.8 `daemon metadata not found`

뜻:

- daemon 상태 파일이 없거나 이미 정리된 상태입니다.

해결:

- `agentab daemon start`
- 또는 세션/탭 명령을 다시 실행해 auto-start 유도

### 2.9 `invalid daemon token`

뜻:

- 저장된 daemon token과 실제 daemon token이 맞지 않습니다.

해결:

```bash
agentab daemon stop
agentab daemon start
```

필요하면 `${AGENTAB_HOME}/run/daemon.json` 상태를 확인합니다.

### 2.10 `lock_conflict`

뜻:

- 다른 owner가 같은 탭에 mutation 작업을 수행 중입니다.

해결:

- 잠시 뒤 다시 시도
- owner 정책을 점검
- concurrent agent가 같은 탭을 공유하고 있지 않은지 확인

## 3. 환경 변수 점검표

문제가 길어질 때 확인할 환경 변수:

- `AGENTAB_HOME`
- `AGENTAB_PINCHTAB_BIN`
- `CHROME_BIN`
- `PINCHTAB_URL`
- `PINCHTAB_TOKEN`
- `AGENTAB_SKIP_INSTALL`

## 4. 상태 파일 위치

확인할 주요 파일:

- `${AGENTAB_HOME}/state.json`
- `${AGENTAB_HOME}/run/daemon.json`
- `${AGENTAB_HOME}/logs/agentab-daemon.log`
- `${AGENTAB_HOME}/logs/pinchtab.log`
- `${AGENTAB_HOME}/artifacts/`

`doctor`에서 바로 확인 가능한 항목:

- `logs`
- `daemon log`
- `pinchtab log`
- `artifacts`

## 5. 재현 순서

문제 재현을 단순하게 하려면 이 순서가 좋습니다.

```bash
agentab --output text doctor
agentab daemon status
agentab session start demo
agentab tab open --session demo https://example.com
agentab tab text --session demo
```

문제가 이 단계 중 어디서 나는지 먼저 좁힌 뒤에, 해당 로그를 확인합니다.

## 6. 그래도 안 되면

함께 남기면 좋은 정보:

- `agentab --output text doctor` 출력
- 실행한 정확한 명령
- `${AGENTAB_HOME}/logs/agentab-daemon.log`
- `${AGENTAB_HOME}/logs/pinchtab.log`
- 사용한 `AGENTAB_HOME`, `CHROME_BIN`, `PINCHTAB_URL`

## 7. 관련 문서

- 설치 및 첫 실행: [cli-install.md](/workspace/agentab-cli/docs/cli-install.md)
- CLI 계약 요약: [cli.md](/workspace/agentab-cli/docs/cli.md)
- CLI 스펙 기준선: [cli-spec-sheet.md](/workspace/agentab-cli/docs/cli-spec-sheet.md)
- 릴리스 체크리스트: [cli-release-checklist.md](/workspace/agentab-cli/docs/cli-release-checklist.md)
