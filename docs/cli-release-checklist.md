# agentab CLI 릴리스 체크리스트

상태: v0.1.1 기준 부분 검증 완료
작성일: 2026-03-17  
마지막 갱신: 2026-03-17 08:30 UTC
목적: `agentab CLI`를 실제 배포 가능한 제품으로 마감하기 전에 확인해야 하는 항목을 표준화하기 위함

## 1. 사용 방법

이 문서는 릴리스 후보마다 처음부터 끝까지 다시 확인하는 체크리스트다.

원칙:

- 체크는 `문서`, `기능`, `신뢰성`, `배포 준비` 순서로 진행한다.
- 가능하면 새 머신 또는 깨끗한 환경에서 다시 확인한다.
- 애매하면 “통과”로 두지 말고 `보류` 또는 `미통과`로 남긴다.

상태 표기:

- `[ ]` 미확인
- `[x]` 확인 완료
- `[~]` 조건부 통과 또는 수동 확인 필요

## 2. 릴리스 후보 정보

릴리스 버전:

- [x] 버전 태그 확정
  값: `v0.1.1`

검증 일시:

- [x] UTC 시각 기록
  값: `2026-03-17 08:30 UTC`

검증 환경:

- [x] OS / 아키텍처 기록
  값: `Linux x86_64`
- [x] Go 버전 기록
  값: `go1.25.0 linux/amd64`
- [x] Chrome 또는 Chromium 버전 기록
  값: `Chrome for Testing 146.0.7680.80`
- [x] PinchTab 버전 기록
  값: `pinchtab 0.8.2`

검증자:

- [x] 담당자 기록
  값: `Codex pair session`

## 3. 문서 준비

- [x] [cli.md](/workspace/agentab-cli/docs/cli.md) 예제가 현재 명령 계약과 맞는다.
- [x] [cli-spec-sheet.md](/workspace/agentab-cli/docs/cli-spec-sheet.md)가 현재 제품 범위와 맞는다.
- [x] 설치 문서 또는 설치 절차가 최신 상태다.
- [x] 트러블슈팅 문서 또는 대표 실패 케이스 정리가 있다.
- [x] artifact 저장 구조 설명이 문서에 반영되어 있다.
- [x] `doctor`, `session`, `tab` 대표 예제가 실제 출력과 맞는다.

## 4. 기본 기능 검증

### 4.1 doctor

- [x] `agentab doctor`가 정상 종료한다.
- [x] `agentabHome`가 올바르게 보인다.
- [x] `artifactsDir`가 올바르게 보인다.
- [x] `managedBinPath`가 올바르게 보인다.
- [x] `pinchtabURL`과 `pinchtabHealthy`가 합리적으로 나온다.
- [~] `chromeBin` 탐지 결과가 현재 환경과 맞다.
  메모: `CHROME_BIN` override로 실제 실행은 되지만 `doctor` 출력의 `chromeBin`은 빈 값으로 보임.

### 4.2 세션과 탭

- [x] `agentab session start demo`
- [x] `agentab tab open --session demo <url>`
- [~] `agentab tab list --session demo`
  메모: 공개 릴리스 asset smoke에서 `tab open` 직후 `about:blank`만 보이는 사례가 한 번 있어 추가 확인 필요.
- [x] `agentab tab text --session demo`
- [x] `agentab tab snapshot --session demo`
- [x] `agentab tab find --session demo "<query>"`
- [x] `agentab tab click --session demo --tab <tabId> --ref <ref>`
- [x] `agentab session stop demo`

### 4.3 artifact

- [x] `agentab tab snapshot --session demo --save`
- [x] `agentab tab snapshot --session demo --out <path>`
- [x] `agentab tab screenshot --session demo --save`
- [x] `agentab tab pdf --session demo --save`
- [x] `${AGENTAB_HOME}/artifacts/...` 아래 파일이 기대한 위치에 생긴다.
- [x] 명시적 `--out` 경로 저장이 동작한다.

## 5. JSON 계약 검증

- [x] 기본 출력이 JSON envelope다.
- [x] 성공 응답에 `ok`, `data`, `diagnostics` 구조가 유지된다.
- [x] 실패 응답에 `error.code`, `error.message`가 유지된다.
- [x] `--output text`가 사람이 읽기 쉬운 결과를 보여준다.
- [x] `--output text`에서도 실패 의미가 손실되지 않는다.

## 6. 종료 코드 검증

- [x] 성공 시 `0`
- [x] 사용법 오류 시 `2`
- [x] 의존성 오류 시 `3`
- [ ] 찾을 수 없음 오류 시 `4`
- [ ] lock conflict 시 `5`
- [ ] timeout 시 `6`
- [ ] 기타 upstream/runtime 오류 시 `7`

## 7. 신뢰성 검증

### 7.1 설치와 부트스트랩

- [~] PinchTab이 없는 환경에서 자동 설치가 재현된다.
  메모: 이전 검증과 worklog에는 기록되어 있으나 `v0.1.1` 공개 asset smoke는 `AGENTAB_PINCHTAB_BIN` override 기준으로 수행함.
- [~] 설치된 PinchTab 바이너리가 `${AGENTAB_HOME}/bin` 아래에 놓인다.
  메모: 자동 설치 경로 자체는 구현되어 있으나 이번 공개 asset smoke에서는 override 경로를 사용함.
- [x] `AGENTAB_PINCHTAB_BIN` override가 동작한다.
- [x] `AGENTAB_HOME` override가 동작한다.

### 7.2 daemon

- [x] daemon auto-start가 동작한다.
- [x] `agentab daemon status`가 현재 상태를 보여준다.
- [x] `agentab daemon stop` 후 상태가 정리된다.
- [ ] 재실행 시 기존 daemon 정보와 충돌하지 않는다.

### 7.3 상태와 복구

- [x] current session이 기대대로 갱신된다.
- [x] current tab이 기대대로 갱신된다.
- [ ] 잘못된 session / tab 지정 시 오류가 일관적이다.
- [ ] daemon 재시작 후에도 상태가 비정상적으로 꼬이지 않는다.

### 7.4 액션 안정성

- [~] 클릭, 입력, 스크롤 기본 동작이 실제 페이지에서 재현된다.
  메모: 클릭은 공개 asset 기준으로 검증했고 입력/스크롤은 이번 릴리스 검증 범위에 포함하지 않았다.
- [ ] lock conflict가 재현되면 올바른 오류 코드와 메시지가 나온다.
- [ ] upstream 오류가 JSON envelope로 표준화된다.

## 8. 테스트 검증

- [x] `go test ./...` 통과
- [x] fake PinchTab 기반 통합 테스트 통과
- [x] live smoke 실행 결과 기록
- [ ] 필요한 경우 headed / headless 각각 확인

## 9. 배포 준비

- [x] [cli-release-workflow.md](/workspace/agentab-cli/docs/cli-release-workflow.md)가 현재 배포 방식과 맞다.
- [x] `.github/workflows/release.yml`이 현재 GoReleaser 설정과 맞다.
- [x] `./scripts/install-goreleaser.sh`가 현재 고정 버전을 설치한다.
- [x] `./scripts/release-snapshot.sh`가 현재 환경에서 성공한다.
- [x] non-git 환경에서는 `snapshot release`만으로 archive와 checksum이 생성됨을 확인했다.
- [x] git 환경에서는 `check + snapshot release`로 동작함을 확인했다.
- [x] 릴리스 노트 초안 작성
- [x] 변경된 공개 계약 정리
- [~] 알려진 제한 사항 정리
  메모: `tab list`와 `doctor.chromeBin` 관련 후속 확인 포인트가 남아 있다.
- [x] 바이너리 배포 방식 결정 또는 검증
- [x] 버전 태그 정책과 실제 태그 값 확인

## 10. 최소 릴리스 게이트

아래 5개를 모두 만족해야 “배포 가능”으로 본다.

- [x] 새 환경에서 `doctor -> session start -> tab open -> text -> click -> screenshot -> session stop`이 문서대로 재현된다.
- [x] `go test ./...`가 통과한다.
- [x] 설치/사용/문제 해결 문서가 현재 구현과 맞다.
- [x] artifact 저장 구조가 실제로 동작한다.
- [x] subprocess로 붙는 소비자 입장에서 JSON 계약이 흔들리지 않는다.

## 11. 릴리스 판정

최종 판정:

- [ ] 배포 가능
- [x] 조건부 배포 가능
- [ ] 배포 보류

메모:

- [x] 남은 이슈 기록
  - 공개 릴리스 asset smoke에서 `tab open` 직후 `tab list`가 `about:blank`만 보이는 사례가 있었다.
  - `doctor`의 `chromeBin`은 현재 `CHROME_BIN` override를 반영하지 않아 빈 값으로 보인다.
  - 첫 실패 태그 `v0.1.0`은 남아 있고, 실제 사용 기준 릴리스는 `v0.1.1`이다.
