# agentab CLI 운영 런북

상태: 초안  
작성일: 2026-03-18  
마지막 갱신: 2026-03-18 03:01 UTC
목적: `agentab-cli`를 운영하거나 장애를 재현할 때, 무엇을 어떤 순서로 확인해야 하는지 한 문서에서 빠르게 찾게 하기 위함

## 1. 가장 먼저 볼 것

장애나 이상 동작이 나오면 아래 순서로 본다.

1. `agentab --output text doctor`
2. `${AGENTAB_HOME}/logs/agentab-daemon.log`
3. `${AGENTAB_HOME}/logs/pinchtab.log`
4. `${AGENTAB_HOME}/artifacts/`

`doctor`에서 바로 확인할 핵심 항목:

- `home`
- `logs`
- `daemon log`
- `pinchtab log`
- `artifacts`
- `chrome`
- `pinchtab`
- `daemon`

## 2. 운영 경로

기본 운영 경로:

- 홈: `${AGENTAB_HOME}` 또는 `${HOME}/.agentab`
- 로그 디렉터리: `${AGENTAB_HOME}/logs`
- daemon 로그: `${AGENTAB_HOME}/logs/agentab-daemon.log`
- PinchTab 로그: `${AGENTAB_HOME}/logs/pinchtab.log`
- artifact 디렉터리: `${AGENTAB_HOME}/artifacts`
- 런타임 상태: `${AGENTAB_HOME}/run`

대표 상태 파일:

- `${AGENTAB_HOME}/state.json`
- `${AGENTAB_HOME}/run/daemon.json`
- `${AGENTAB_HOME}/run/pinchtab.json`

## 3. 첫 진단 순서

### 3.1 브라우저가 안 뜰 때

1. `agentab --output text doctor`
2. `chrome` 섹션에서 `status`, `source`, `path` 확인
3. `pinchtab` 섹션에서 `health`, `url`, `binary error` 확인
4. `pinchtab.log` 확인

### 3.2 세션/탭 명령이 실패할 때

1. `agentab daemon status`
2. `state.json`에서 current session / current tab 확인
3. `agentab-daemon.log` 확인
4. 실패 응답의 `error.code`와 종료 코드 확인

### 3.3 artifact가 기대와 다를 때

1. 명령 응답의 `path` 확인
2. `managed=true/false` 확인
3. 관리형이면 `relativePath` 확인
4. `createdAt`으로 최근 생성 여부 확인

## 4. 운영 smoke

기본 모드/action smoke:

```bash
cd /workspace/agentab-cli
AGENTAB_BIN=/workspace/agentab-cli/agentab \
./scripts/smoke-modes.sh
```

이 스크립트는 아래를 검증한다.

- `headless`
- `headed`
- `click`
- `type`
- `fill`
- `press`
- `scroll`

display가 없는 Linux에서는 `Xvfb`가 필요하다.

## 5. 이슈에 첨부하면 좋은 정보

- `agentab --output text doctor` 출력
- 실행한 정확한 명령
- `${AGENTAB_HOME}/logs/agentab-daemon.log`
- `${AGENTAB_HOME}/logs/pinchtab.log`
- 관련 artifact 경로와 파일
- 사용한 `AGENTAB_HOME`, `CHROME_BIN`, `PINCHTAB_URL`

## 6. 릴리스 검증 문서

릴리스별 검증 이력:

- [release verification index](/workspace/agentab-cli/docs/releases/README.md)

대표 릴리스 검증 문서:

- [v0.1.1 verification](/workspace/agentab-cli/docs/releases/v0.1.1-verification.md)
- [v0.1.2 verification](/workspace/agentab-cli/docs/releases/v0.1.2-verification.md)
- [v0.1.3 verification](/workspace/agentab-cli/docs/releases/v0.1.3-verification.md)
