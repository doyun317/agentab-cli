# agentab CLI 릴리스 체크리스트

상태: 초안  
작성일: 2026-03-17  
마지막 갱신: 2026-03-17 06:12 UTC  
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

- [ ] 버전 태그 확정

검증 일시:

- [ ] UTC 시각 기록

검증 환경:

- [ ] OS / 아키텍처 기록
- [ ] Go 버전 기록
- [ ] Chrome 또는 Chromium 버전 기록
- [ ] PinchTab 버전 기록

검증자:

- [ ] 담당자 기록

## 3. 문서 준비

- [ ] [cli.md](/workspace/agentab-cli/docs/cli.md) 예제가 현재 명령 계약과 맞는다.
- [ ] [cli-spec-sheet.md](/workspace/agentab-cli/docs/cli-spec-sheet.md)가 현재 제품 범위와 맞는다.
- [ ] 설치 문서 또는 설치 절차가 최신 상태다.
- [ ] 트러블슈팅 문서 또는 대표 실패 케이스 정리가 있다.
- [ ] artifact 저장 구조 설명이 문서에 반영되어 있다.
- [ ] `doctor`, `session`, `tab` 대표 예제가 실제 출력과 맞는다.

## 4. 기본 기능 검증

### 4.1 doctor

- [ ] `agentab doctor`가 정상 종료한다.
- [ ] `agentabHome`가 올바르게 보인다.
- [ ] `artifactsDir`가 올바르게 보인다.
- [ ] `managedBinPath`가 올바르게 보인다.
- [ ] `pinchtabURL`과 `pinchtabHealthy`가 합리적으로 나온다.
- [ ] `chromeBin` 탐지 결과가 현재 환경과 맞다.

### 4.2 세션과 탭

- [ ] `agentab session start demo`
- [ ] `agentab tab open --session demo <url>`
- [ ] `agentab tab list --session demo`
- [ ] `agentab tab text --session demo`
- [ ] `agentab tab snapshot --session demo`
- [ ] `agentab tab find --session demo "<query>"`
- [ ] `agentab tab click --session demo --tab <tabId> --ref <ref>`
- [ ] `agentab session stop demo`

### 4.3 artifact

- [ ] `agentab tab snapshot --session demo --save`
- [ ] `agentab tab snapshot --session demo --out <path>`
- [ ] `agentab tab screenshot --session demo --save`
- [ ] `agentab tab pdf --session demo --save`
- [ ] `${AGENTAB_HOME}/artifacts/...` 아래 파일이 기대한 위치에 생긴다.
- [ ] 명시적 `--out` 경로 저장이 동작한다.

## 5. JSON 계약 검증

- [ ] 기본 출력이 JSON envelope다.
- [ ] 성공 응답에 `ok`, `data`, `diagnostics` 구조가 유지된다.
- [ ] 실패 응답에 `error.code`, `error.message`가 유지된다.
- [ ] `--output text`가 사람이 읽기 쉬운 결과를 보여준다.
- [ ] `--output text`에서도 실패 의미가 손실되지 않는다.

## 6. 종료 코드 검증

- [ ] 성공 시 `0`
- [ ] 사용법 오류 시 `2`
- [ ] 의존성 오류 시 `3`
- [ ] 찾을 수 없음 오류 시 `4`
- [ ] lock conflict 시 `5`
- [ ] timeout 시 `6`
- [ ] 기타 upstream/runtime 오류 시 `7`

## 7. 신뢰성 검증

### 7.1 설치와 부트스트랩

- [ ] PinchTab이 없는 환경에서 자동 설치가 재현된다.
- [ ] 설치된 PinchTab 바이너리가 `${AGENTAB_HOME}/bin` 아래에 놓인다.
- [ ] `AGENTAB_PINCHTAB_BIN` override가 동작한다.
- [ ] `AGENTAB_HOME` override가 동작한다.

### 7.2 daemon

- [ ] daemon auto-start가 동작한다.
- [ ] `agentab daemon status`가 현재 상태를 보여준다.
- [ ] `agentab daemon stop` 후 상태가 정리된다.
- [ ] 재실행 시 기존 daemon 정보와 충돌하지 않는다.

### 7.3 상태와 복구

- [ ] current session이 기대대로 갱신된다.
- [ ] current tab이 기대대로 갱신된다.
- [ ] 잘못된 session / tab 지정 시 오류가 일관적이다.
- [ ] daemon 재시작 후에도 상태가 비정상적으로 꼬이지 않는다.

### 7.4 액션 안정성

- [ ] 클릭, 입력, 스크롤 기본 동작이 실제 페이지에서 재현된다.
- [ ] lock conflict가 재현되면 올바른 오류 코드와 메시지가 나온다.
- [ ] upstream 오류가 JSON envelope로 표준화된다.

## 8. 테스트 검증

- [ ] `go test ./...` 통과
- [ ] fake PinchTab 기반 통합 테스트 통과
- [ ] live smoke 실행 결과 기록
- [ ] 필요한 경우 headed / headless 각각 확인

## 9. 배포 준비

- [ ] [cli-release-workflow.md](/workspace/agentab-cli/docs/cli-release-workflow.md)가 현재 배포 방식과 맞다.
- [ ] `.github/workflows/release.yml`이 현재 GoReleaser 설정과 맞다.
- [ ] `./scripts/install-goreleaser.sh`가 현재 고정 버전을 설치한다.
- [ ] `./scripts/release-snapshot.sh`가 현재 환경에서 성공한다.
- [ ] non-git 환경에서는 `snapshot release`만으로 archive와 checksum이 생성됨을 확인했다.
- [ ] git 환경에서는 `check + snapshot release`로 동작함을 확인했다.
- [ ] 릴리스 노트 초안 작성
- [ ] 변경된 공개 계약 정리
- [ ] 알려진 제한 사항 정리
- [ ] 바이너리 배포 방식 결정 또는 검증
- [ ] 버전 태그 정책과 실제 태그 값 확인

## 10. 최소 릴리스 게이트

아래 5개를 모두 만족해야 “배포 가능”으로 본다.

- [ ] 새 환경에서 `doctor -> session start -> tab open -> text -> click -> screenshot -> session stop`이 문서대로 재현된다.
- [ ] `go test ./...`가 통과한다.
- [ ] 설치/사용/문제 해결 문서가 현재 구현과 맞다.
- [ ] artifact 저장 구조가 실제로 동작한다.
- [ ] subprocess로 붙는 소비자 입장에서 JSON 계약이 흔들리지 않는다.

## 11. 릴리스 판정

최종 판정:

- [ ] 배포 가능
- [ ] 조건부 배포 가능
- [ ] 배포 보류

메모:

- [ ] 남은 이슈 기록
