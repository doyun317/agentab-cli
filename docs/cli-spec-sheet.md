# agentab CLI 스펙 시트

상태: CLI 본체 기준 베이스라인  
최종 업데이트: 2026-03-17  
문서 목적: `agentab`을 LangChain이나 특정 에이전트 프레임워크와 분리된 독립 제품으로 정의하고, CLI 본체의 공개 계약과 비목표를 고정하기 위함

## 1. 제품 정의

`agentab`은 에이전트와 사람이 공통으로 사용할 수 있는 로컬 우선 브라우저 조작 CLI입니다.

이 제품의 핵심 목표는 다음과 같습니다.

- 브라우저 제어를 위한 단일 진입점 제공
- 세션/탭 중심 상태 관리 표준화
- PinchTab과 Chrome 런타임의 복잡성을 CLI 뒤로 숨기기
- 기계 친화적인 JSON 계약 유지
- 에이전트가 직접 브라우저 드라이버를 구현하지 않도록 만들기

한 줄 정의:

“`agentab`은 에이전트가 브라우저를 안정적으로 조작하기 위해 사용하는 배포 가능한 브라우저 제어 CLI다.”

## 2. 제품 범위

### 2.1 포함 범위

- 로컬 CLI 바이너리 `agentab`
- 로컬 daemon 자동 기동과 상태 관리
- PinchTab 설치 확인, 자동 설치, 자동 실행
- session / tab / lock 관리
- 브라우저 primitive 명령
- JSON envelope와 종료 코드
- 로컬 상태 저장
- low-level trace / artifact 기반

### 2.2 비목표

- LangChain 전용 도구나 planner runtime 내장
- 특정 모델용 prompt 최적화
- 에이전트 workflow orchestration
- 내장 멀티에이전트 시스템
- 원격 멀티테넌트 브라우저 제어 평면

## 3. 대상 사용자

주요 사용자:

- 브라우저 상호작용이 필요한 LLM 에이전트를 만드는 개발자
- shell, Python, TypeScript, Go 등에서 CLI를 subprocess로 붙이고 싶은 개발자
- 브라우저 상태를 직접 CLI로 점검하고 싶은 운영자 또는 개발자

핵심 사용 시나리오:

- 새 세션 생성
- 페이지 열기
- 페이지 텍스트 읽기
- 요소 찾기
- 클릭/입력/스크롤
- 스크린샷 또는 PDF 저장
- 문제 발생 시 `doctor`와 JSON 오류로 원인 파악

## 4. 상위 구조

```text
사용자 또는 에이전트
  -> agentab CLI
  -> agentab daemon
  -> PinchTab
  -> Chrome/Chromium
```

구성 요소:

- CLI entrypoint: `cmd/agentab/main.go`
- command dispatcher: `internal/app/app.go`
- daemon client/server: `internal/daemon/`
- PinchTab install/manager/client: `internal/install/`, `internal/pinchtab/`
- local state store: `internal/state/`
- response envelope: `internal/response/response.go`

## 5. 공개 인터페이스

### 5.1 바이너리

- 공개 바이너리 이름: `agentab`

### 5.2 전역 플래그

- `--session`
- `--tab`
- `--profile`
- `--mode`
- `--owner`
- `--timeout`
- `--output`
- `--debug`

기본값:

- `--timeout`: `30s`
- `--output`: `json`

### 5.3 최상위 명령 그룹

- `doctor`
- `daemon`
- `session`
- `tab`

### 5.4 주요 서브커맨드

`doctor`

- `agentab doctor`

`daemon`

- `agentab daemon start`
- `agentab daemon status`
- `agentab daemon stop`

`session`

- `agentab session start <name>`
- `agentab session list`
- `agentab session resume <name>`
- `agentab session stop [name]`

`tab`

- `agentab tab open --session <name> <url>`
- `agentab tab list --session <name>`
- `agentab tab close --session <name> --tab <tabId>`
- `agentab tab focus --session <name> --tab <tabId>`
- `agentab tab snapshot --session <name> --tab <tabId>`
- `agentab tab text --session <name> --tab <tabId>`
- `agentab tab find --session <name> --tab <tabId> "<query>"`
- `agentab tab click --session <name> --tab <tabId> --ref <ref>`
- `agentab tab type --session <name> --tab <tabId> --ref <ref> "<text>"`
- `agentab tab press --session <name> --tab <tabId> --key <key>`
- `agentab tab scroll --session <name> --tab <tabId>`
- `agentab tab screenshot --session <name> --tab <tabId>`
- `agentab tab pdf --session <name> --tab <tabId>`

## 6. JSON 계약

모든 기계 친화적 명령의 기본 응답 형태는 아래와 같습니다.

```json
{
  "ok": true,
  "data": {},
  "error": null,
  "diagnostics": {}
}
```

필드 의미:

- `ok`: 성공 여부
- `data`: 명령 결과 payload
- `error`: `code`, `message`, 선택적 `details`
- `diagnostics`: 추가 런타임 메타데이터

텍스트 출력 정책:

- `--output json`은 전체 envelope를 출력한다.
- `--output text`는 성공 시 사람이 읽기 쉬운 결과를 출력한다.
- `--output text`에서도 실패는 구조화된 에러 의미를 보존한다.

## 7. 오류 코드와 종료 코드

오류 코드는 envelope의 `error.code`에 담긴다.

현재 종료 코드 매핑:

- `0`: 성공
- `2`: `usage_error`
- `3`: `dependency_error`
- `4`: `not_found`
- `5`: `lock_conflict`
- `6`: `timeout`
- `7`: 그 외 runtime 또는 upstream 오류

이 규칙은 스크립트, 에이전트 래퍼, CI에서 공통으로 사용한다.

## 8. 상태 모델

### 8.1 세션

세션은 사람이 읽기 쉬운 이름을 PinchTab 인스턴스에 매핑하는 로컬 단위다.

주요 필드:

- `name`
- `instanceId`
- `profileId`
- `mode`
- `lastUsedAt`
- `currentTabId`

### 8.2 daemon

daemon은 로컬 loopback HTTP 서버이며, CLI가 auto-start 할 수 있다.

주요 필드:

- `port`
- `token`
- `pid`
- `startedAt`

### 8.3 tab

tab은 PinchTab tab ID와 로컬 세션 이름을 연결하는 조작 대상이다.

주요 사용 속성:

- `tabId`
- `url`
- `title`
- 현재 focus 상태
- lock owner / lock ttl

## 9. 로컬 저장소와 경로

기본 루트:

- `${HOME}/.agentab`
- `AGENTAB_HOME`으로 override 가능

중요 경로:

- `${AGENTAB_HOME}/state.json`
- `${AGENTAB_HOME}/run/daemon.json`
- `${AGENTAB_HOME}/logs/`
- `${AGENTAB_HOME}/artifacts/`
- `${AGENTAB_HOME}/bin/`

책임:

- `state.json`: 세션과 current session 저장
- `run/daemon.json`: daemon port/token/pid 저장
- `logs/`: 로그와 추가 runtime 산출물 보관 여지
- `artifacts/`: screenshot, snapshot, pdf 같은 CLI 산출물 저장
- `bin/`: 관리형 PinchTab 바이너리 저장

## 10. PinchTab 통합

PinchTab은 `agentab`의 브라우저 런타임 백엔드다.

바이너리 해석 순서:

1. `AGENTAB_PINCHTAB_BIN`
2. `PATH` 상의 `pinchtab`
3. `${AGENTAB_HOME}/bin` 아래 관리형 바이너리
4. 없으면 최신 GitHub release 다운로드 및 설치

실행 정책:

- 기본 URL은 `http://127.0.0.1:9867`
- `PINCHTAB_URL`이 지정되면 우선 사용
- 로컬 URL이고 접근 실패 시 자동 설치/자동 실행을 시도
- `CHROME_BIN`이 있으면 PinchTab config에 `browser.binary`로 전달하고, `doctor`는 해당 경로를 `chromeBin`과 `chromeBinSource=env`로 보여준다

## 11. daemon 정책

daemon 기본 정책:

- bind 주소: `127.0.0.1`
- 인증: 로컬 상태에 저장된 bearer token
- 포트: 기존 기록 우선, 실패 시 사용 가능한 loopback 포트 선택
- 역할:
  - session/tab 라우트 제공
  - PinchTab subprocess 생명주기 소유
  - tab mutation lock 처리
  - PinchTab 응답을 CLI 계약으로 표준화

## 12. 핵심 제품 원칙

### 12.1 primitive-first

CLI는 상위 workflow가 아니라 stable primitive를 제공한다.

예:

- `open`
- `text`
- `find`
- `click`
- `type`
- `scroll`
- `screenshot`

### 12.2 framework-independent

CLI 본체는 LangChain, CrewAI, 특정 SDK를 몰라도 동작해야 한다.

### 12.3 local-first

기본 사용성은 로컬 단일 머신 기준으로 설계한다.

### 12.4 JSON-first

에이전트와 스크립트가 파싱하기 쉽게 JSON 계약을 기본값으로 유지한다.

## 13. 현재 출시 범위

현재 제품 범위에 포함되는 것으로 보는 항목:

- `doctor`, `daemon`, `session`, `tab` 기본 명령 체계
- PinchTab 자동 설치와 자동 실행
- session/tab 상태 저장
- 기본 browser primitive 조작
- managed artifact 저장 구조
- JSON 응답 계약

현재 제품 범위에서 제외하는 항목:

- planner agent
- LangChain tool wrapper
- benchmark harness
- prompt tuning
- model-specific recovery

이들은 모두 CLI 바깥의 통합 또는 실험 계층으로 본다.

## 14. 출시 완료 기준

`agentab CLI`를 배포 가능한 제품으로 보기 위한 최소 기준:

- 새 머신에서 `doctor -> session start -> tab open -> text -> click -> screenshot -> session stop`이 문서대로 재현된다.
- fake PinchTab 기반 통합 테스트가 안정적으로 통과한다.
- live smoke가 최소 기준 페이지에서 재현된다.
- 설치/문제 해결/명령 예제가 문서로 정리되어 있다.
- 에이전트가 subprocess로 붙일 때 계약이 흔들리지 않는다.

## 15. 이 문서의 사용 규칙

앞으로 CLI 본체 작업은 먼저 이 문서를 기준으로 판단한다.

판단 질문:

1. 이 변경이 CLI의 공개 계약을 바꾸는가?
2. 이 변경이 framework-independent하게 유용한가?
3. 이 변경이 본체 제품 안정성을 올리는가?

셋 중 둘 이상이 맞으면 CLI 트랙 작업으로 간주한다.
