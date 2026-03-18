# agentab CLI 로드맵 및 작업 기록

상태: CLI 본체 전용 운영 문서  
최초 작성: 2026-03-17  
마지막 갱신: 2026-03-18 04:36 UTC
문서 목적: `agentab CLI` 본체 제품의 구현 로드맵, 작업 우선순위, 변경 기록, 출시 기준을 LangChain 트랙과 분리해 관리하기 위함

## 1. 이 문서의 목적

이 문서는 `agentab CLI` 본체만을 위한 작업 노트다.

이 문서에서 관리하는 것:

- CLI 제품 로드맵
- 공개 계약 안정화 작업
- 설치/실행/운영성 개선
- CLI 전용 변경 기록
- CLI 출시 준비 상태

이 문서에서 다루지 않는 것:

- LangChain benchmark 세부 튜닝
- planner / executor 실험
- 모델별 프롬프트 최적화

## 2. 현재 CLI 제품 정의

한 줄 정의:

“`agentab`은 에이전트가 브라우저를 조작하기 위해 사용하는 배포 가능한 로컬 브라우저 제어 CLI다.”

현재 제품 위치:

- Go 기반 CLI/daemon은 이미 존재한다.
- PinchTab 설치/실행/상태 관리 경로도 존재한다.
- session/tab 모델과 JSON envelope 계약도 존재한다.
- 따라서 지금 단계는 “아이디어 검증”이 아니라 “CLI 제품 마감” 단계에 가깝다.

## 3. 현재 상태 요약

이미 확보된 것:

- `doctor`, `daemon`, `session`, `tab` 명령 체계
- JSON envelope와 종료 코드
- PinchTab 자동 설치
- local daemon 자동 기동
- session / tab 상태 저장
- PinchTab subprocess 관리
- 기본 browser primitive

아직 마감이 필요한 것:

- 후속 patch release 범위 결정
- 액션 smoke를 다음 patch release 검증 기록에 반영
- 공개 release asset 기준 액션 smoke 승격 여부 결정

## 4. 단계별 로드맵

## 4.1 Phase A: 계약 고정

목표:

- CLI 제품의 공개 계약을 흔들리지 않게 고정한다.

작업:

- `doctor`, `daemon`, `session`, `tab` 명령 계약 점검
- JSON envelope와 `--output text` 동작 점검
- 종료 코드 표준화 문서화
- 글로벌 플래그 의미 확정

완료 기준:

- 에이전트와 사람이 모두 같은 명령 계약을 문서대로 사용할 수 있다.

## 4.2 Phase B: 런타임 신뢰성

목표:

- 새 머신과 재시작 상황에서도 안정적으로 동작하게 만든다.

작업:

- auto-start daemon 안정성 점검
- PinchTab 설치/업데이트 경로 안정화
- Chrome binary 확인 로직 보강
- timeout / upstream error / lock conflict 처리 보강
- 상태 복구와 재시작 동작 점검

완료 기준:

- 초기 실행과 재실행이 예측 가능하고, 실패 시 원인 파악이 쉽다.

## 4.3 Phase C: 관측성과 산출물

목표:

- CLI 사용 중 생기는 산출물을 제품 관점에서 정리한다.

작업:

- screenshot 저장 구조 정리
- snapshot artifact 저장 구조 정리
- trace / manifest / report의 CLI 책임 범위 정리
- 기본 logs / artifact 경로 정책 정리

완료 기준:

- CLI 단독 사용만으로도 실패 상황을 다시 볼 수 있는 최소 artifact 흐름이 있다.

## 4.4 Phase D: 문서와 배포

목표:

- CLI를 실제 제품처럼 배포할 수 있게 만든다.

작업:

- 설치 문서 정리
- 사용 예제 정리
- 트러블슈팅 정리
- release checklist 작성
- 배포 형식 결정

완료 기준:

- 처음 보는 사용자도 문서만 보고 설치와 첫 실행이 가능하다.

## 4.5 Phase E: 출시 준비

목표:

- CLI를 반복 가능하게 릴리스할 수 있는 상태로 만든다.

작업:

- smoke 시나리오 확정
- 테스트 매트릭스 정리
- 버전 정책 확정
- 변경 로그 기준 정리

완료 기준:

- 릴리스 전 검증 절차가 문서화되어 있고 반복 가능하다.

## 5. 현재 활성 작업

상태 구분:

- `todo`
- `doing`
- `blocked`
- `done`
- `dropped`

### 5.1 계약과 핵심 기능

- `done` CLI 명령 골격 구현
- `done` JSON envelope 계약 구현
- `done` 종료 코드 매핑 구현
- `done` session / tab 기본 명령 구현
- `doing` 명령별 text output 경험 점검
- `todo` CLI 예제 명령 셋 정리
- `done` 종료 코드 테스트 보강

### 5.2 런타임과 신뢰성

- `done` PinchTab 자동 설치 경로 구현
- `done` daemon auto-start 경로 구현
- `done` PinchTab / daemon child process detach 보완
- `done` 최신 PinchTab의 `browser.binary` 전달 보완
- `done` 멀티 인스턴스 `tab list` 경로 수정
- `done` 완전 새 환경 기준 PinchTab 자동 설치 smoke 재검증
- `done` 잘못된 session / tab / 종료 코드 검증 보강
- `done` daemon shutdown 시 runtime/session state 정리
- `done` headed / headless 이중 smoke
- `todo` lock / timeout / upstream error 문서화 및 검증
- `done` auto-install / daemon cleanup fix 포함 next patch release 발행

### 5.3 운영성과 산출물

- `done` screenshot / snapshot artifact 저장 구조 완성
- `todo` CLI 기준 trace 책임 범위 확정
- `todo` logs / artifact 기본 경로 정책 정리
- `todo` doctor 결과에 artifact 경로 진단 포함 여부 검토

### 5.4 문서와 출시

- `done` 기본 CLI 문서 초안 작성
- `done` 설치 문서 강화
- `done` 트러블슈팅 문서 추가
- `done` release checklist 초안 작성
- `done` 배포 방식 결정
- `done` GoReleaser local snapshot workflow 구현
- `done` git 저장소 생성 후 GitHub Actions release workflow 연결
- `done` `v0.1.2` patch release 발행
- `todo` `v0.1.0` 실패 태그 처리 여부 결정
- `todo` 릴리스/체크리스트 문서 중복 정리

## 6. 현재 추천 1순위

- `doctor --output text` 가독성 점검

이유:

- 릴리스 기준 기능/모드 smoke는 상당 부분 닫혔고, 지금 남은 가장 직접적인 사용자 경험 개선은 사람이 직접 읽는 `doctor --output text`의 품질을 점검하는 것이기 때문

## 7. 출시 체크리스트 초안

### 7.1 기능 검증

- [ ] `doctor`가 핵심 환경 정보를 정확히 보여준다.
- [ ] `session start -> tab open -> text -> click -> screenshot -> session stop` 흐름이 통과한다.
- [ ] headless / headed 기본 모드가 모두 검증된다.

### 7.2 신뢰성 검증

- [ ] PinchTab 자동 설치가 새 환경에서 재현된다.
- [ ] daemon 재기동 후 상태가 정리된다.
- [ ] 잘못된 세션/탭/락 충돌 시 오류가 일관적이다.

### 7.3 문서 검증

- [ ] 설치 문서만 보고 첫 실행이 가능하다.
- [ ] 문제 해결 문서가 대표 실패 케이스를 포함한다.
- [ ] JSON 예제가 현재 실제 출력과 맞다.

## 8. CLI 기준선 기록

## 8.1 기준선 A

기록일:

- 2026-03-17

설명:

- CLI 골격, daemon, session/tab, PinchTab 자동 설치 경로가 구현된 상태를 CLI 제품의 첫 기준선으로 본다.

의미:

- 이 시점부터 CLI는 실험 코드가 아니라 제품 마감 대상으로 관리한다.

## 8.2 기준선 B

기록일:

- 2026-03-17

설명:

- 실제 머신에서 PinchTab 자동 다운로드, 세션 생성, 브라우저 실행 경로를 검증하면서 child process context 문제와 browser binary 전달 문제를 수정했다.

의미:

- CLI 런타임 신뢰성에서 중요한 실제 버그들이 이미 한 차례 정리되었다.

## 8.3 기준선 C

기록일:

- 2026-03-17

설명:

- live browser smoke에서 `session start -> tab open -> tab list -> tab text -> tab snapshot -> tab find -> tab click -> session stop`까지 확인했다.

의미:

- CLI는 최소한의 end-to-end 조작 제품으로서 동작 가능하다는 실증이 있다.

## 9. 주요 결정 기록

### 9.1 제품 포지션

- 결정: `agentab` 본체는 browser control CLI로 정의한다.
- 이유: LangChain 같은 상위 프레임워크와 분리되어도 가치가 유지되는 제품이어야 하기 때문이다.

### 9.2 인터페이스

- 결정: 기본 출력은 JSON envelope로 유지한다.
- 이유: 에이전트와 스크립트가 예측 가능하게 파싱할 수 있어야 하기 때문이다.

### 9.3 런타임

- 결정: PinchTab은 브라우저 런타임 백엔드로 사용한다.
- 이유: 브라우저 엔진을 재구현하지 않고 CLI 본체의 운영성과 계약에 집중하기 위해서다.

### 9.4 경계

- 결정: planner / benchmark / prompt tuning은 CLI 본체가 아니라 통합 계층으로 둔다.
- 이유: CLI 본체를 framework-independent 제품으로 유지하기 위해서다.

## 10. 변경 로그

### 2026-03-17 05:55 UTC

변경:

- CLI 본체만을 위한 별도 스펙시트와 로드맵/작업기록 문서를 분리했다.
- 앞으로 CLI 관련 의사결정과 우선순위 관리는 이 문서를 기준으로 진행한다.

이유:

- LangChain 통합 트랙과 CLI 본체 제품 트랙을 섞지 않고, 본체 제품 마감을 더 분명하게 하기 위해

영향:

- CLI 작업은 이제 `docs/cli-spec-sheet.md`와 이 문서를 기준으로 판단한다.
- adapter 실험과 제품 본체 작업의 경계가 더 분명해졌다.

후속 작업:

- screenshot / snapshot artifact 저장 구조 완성
- release checklist 초안 작성
- 설치/실행/트러블슈팅 문서 보강

### 2026-03-17 06:04 UTC

변경:

- `${AGENTAB_HOME}/artifacts` 루트를 CLI 저장소 레이아웃에 추가함
- `tab snapshot`에 `--out`, `--save`를 추가해 snapshot artifact를 파일로 저장할 수 있게 함
- `tab screenshot`, `tab pdf`에 `--save`를 추가하고 관리형 artifact 경로에 기본 저장할 수 있게 함
- `doctor` 결과에 `artifactsDir`를 포함함

이유:

- CLI를 제품처럼 마무리하려면 사용자가 파일 artifact를 다시 열어보고 디버깅할 수 있는 일관된 저장 구조가 필요했기 때문

영향:

- screenshot / snapshot / pdf 산출물을 `${AGENTAB_HOME}/artifacts/<session>/<tab>/...` 아래에 정리할 수 있게 됨
- 명시적 `--out`이 없더라도 `--save`만으로 관리형 artifact 저장이 가능해짐
- CLI 제품의 산출물 경계가 logs와 분리되어 더 선명해짐

후속 작업:

- CLI release checklist 초안 작성
- 설치/실행/트러블슈팅 문서 보강
- CLI 기준 trace 책임 범위 확정

### 2026-03-17 06:12 UTC

변경:

- CLI 릴리스 전 검증 항목을 별도 문서 [cli-release-checklist.md](/workspace/agentab-cli/docs/cli-release-checklist.md)로 분리함
- worklog 안에 있던 짧은 출시 체크리스트 초안을 독립 문서 기준으로 승격함

이유:

- CLI를 제품처럼 마무리하려면 릴리스 직전 검증 항목이 worklog 안의 메모가 아니라 실제 운영 문서여야 하기 때문

영향:

- 이제 CLI 배포 준비는 [cli-release-checklist.md](/workspace/agentab-cli/docs/cli-release-checklist.md)를 기준으로 진행한다.
- worklog의 다음 우선순위가 문서 보강으로 자연스럽게 이동했다.

후속 작업:

- 설치/실행/트러블슈팅 문서 보강
- 배포 방식 결정
- CLI 기준 trace 책임 범위 확정

### 2026-03-17 06:19 UTC

변경:

- CLI 전용 설치 및 첫 실행 가이드 [cli-install.md](/workspace/agentab-cli/docs/cli-install.md)를 추가함
- CLI 전용 트러블슈팅 가이드 [cli-troubleshooting.md](/workspace/agentab-cli/docs/cli-troubleshooting.md)를 추가함
- README와 CLI 개요 문서에서 새 가이드들로 연결되는 링크를 보강함

이유:

- 릴리스 체크리스트의 문서 준비 항목을 실제 사용자 문서로 채우고, 첫 사용자와 운영자가 바로 참조할 수 있는 경로를 만들어야 했기 때문

영향:

- 설치, 첫 실행, artifact 저장 위치, 로그 경로, 대표 실패 메시지 대응이 문서화됨
- CLI 제품 문서의 최소 세트가 `개요 + 설치 + 트러블슈팅 + 릴리스 체크리스트` 형태로 갖춰짐
- worklog의 다음 우선순위가 배포 방식 결정으로 이동함

후속 작업:

- 배포 방식 결정
- CLI 기준 trace 책임 범위 확정
- JSON/text output 예제 보강

### 2026-03-17 06:47 UTC

변경:

- 배포 방식을 `GitHub Releases + GoReleaser`로 고정했다.
- 로컬 배포 준비용 `.goreleaser.yaml`을 추가했다.
- `scripts/install-goreleaser.sh`, `scripts/release-snapshot.sh`를 추가해 git 저장소가 없을 때도 snapshot build를 검증할 수 있게 했다.
- 배포 준비 흐름을 [cli-release-workflow.md](/workspace/agentab-cli/docs/cli-release-workflow.md)로 분리했다.

이유:

- 아직 GitHub 저장소가 없더라도, 릴리스 방식과 빌드 기준을 먼저 고정해두면 저장소 생성 이후에 바로 태그 릴리스 자동화로 넘어갈 수 있기 때문

영향:

- 이제 `/workspace/agentab-cli`에서 GoReleaser 설정과 cross-platform snapshot build를 로컬에서 반복 검증할 수 있다.
- 실제 GitHub Release 업로드와 Actions workflow는 git 저장소 생성 이후 단계로 명확히 분리되었다.
- CLI 릴리스 체크리스트에 snapshot script 검증 항목이 추가되었다.
- `goreleaser check`는 git 저장소 바깥에서 실패하므로, 현재 non-git 경로에서는 snapshot build 자체를 검증 단계로 사용한다.

후속 작업:

- git 저장소 생성 후 GitHub Actions release workflow 연결
- 릴리스 노트 초안 작성
- JSON/text output 예제 보강

### 2026-03-17 06:56 UTC

변경:

- non-git 환경에서도 `goreleaser release --snapshot --clean`이 성공하고 archive와 checksum까지 생성되는 것을 확인했다.
- `release-snapshot.sh`를 binary-only build 경로에서 snapshot release 경로로 바꿨다.
- 배포 문서에 실제 사용자-facing archive 이름 예시를 추가했다.

이유:

- 내부 `dist/..._v1`, `..._v8.0` 경로보다 실제 릴리스 asset 이름이 사용자 관점에서 더 중요하고, non-git 환경에서도 그것을 바로 검증할 수 있기 때문

영향:

- 이제 `/workspace/agentab-cli/dist`에서 clean archive 이름을 바로 확인할 수 있다.
- 내부 build 디렉터리와 사용자-facing release asset 이름의 역할이 문서상으로도 분리되었다.
- 앞으로는 non-git 환경에서도 release naming 기준을 거의 실제와 동일하게 검증할 수 있다.

후속 작업:

- git 저장소 생성 후 GitHub Actions release workflow 연결
- 릴리스 노트 초안 작성
- JSON/text output 예제 보강

### 2026-03-17 06:58 UTC

변경:

- GoReleaser snapshot 버전 템플릿을 `0.1.0-snapshot`으로 올렸다.
- 배포 워크플로우 문서의 예시 archive 이름도 `0.1.0-snapshot` 기준으로 갱신했다.

이유:

- 첫 공개 릴리스 후보를 `0.1.0`으로 보고 있으므로, 로컬 snapshot 검증 산출물도 같은 버전 축을 쓰는 편이 제품 기준선과 더 잘 맞기 때문

영향:

- `dist/`에 생성되는 snapshot archive 이름이 이제 `agentab-cli_0.1.0-snapshot_*` 형태로 나온다.
- 문서 예시와 실제 산출물 버전 기준이 일치한다.

후속 작업:

- git 저장소 생성 후 GitHub Actions release workflow 연결
- 릴리스 노트 초안 작성
- JSON/text output 예제 보강

### 2026-03-17 07:07 UTC

변경:

- `AGENTAB_HOME`이 다른 릴리스 바이너리가 기존 `127.0.0.1:9867` PinchTab을 재사용하던 문제를 수정했다.
- PinchTab runtime 주소를 홈별 `run/pinchtab.json`에 저장하고, 기본 포트가 점유 중이면 다른 로컬 포트를 선택하도록 바꿨다.
- `doctor`도 같은 runtime 선택 로직을 사용하도록 맞췄다.

이유:

- 배포 아카이브를 새 폴더에 풀어 검증할 때 기존 글로벌 PinchTab runtime에 기대면, CLI가 독립적으로 배포 가능한지 제대로 증명할 수 없기 때문

영향:

- 서로 다른 `AGENTAB_HOME`은 PinchTab 포트까지 격리된 런타임을 사용할 수 있다.
- 기존 `9867`이 살아 있어도 새 릴리스 바이너리가 다른 포트에서 자체 PinchTab을 띄운다.
- 패키지된 아카이브를 새 폴더에 풀어 `doctor -> session start -> tab open -> text -> click -> screenshot -> daemon stop`까지 독립적으로 검증했다.

후속 작업:

- git 저장소 생성 후 GitHub Actions release workflow 연결
- 릴리스 노트 초안 작성
- JSON/text output 예제 보강

### 2026-03-17 08:10 UTC

변경:

- GitHub 저장소 연결 이후 tag 기반 릴리스를 위한 GitHub Actions workflow 파일 `.github/workflows/release.yml`을 추가했다.
- `v*` 태그 push 시 GoReleaser가 release를 수행하도록 설정했다.
- 배포 워크플로우 문서와 릴리스 체크리스트를 현재 상태에 맞게 갱신했다.

이유:

- 로컬 snapshot 검증만으로는 실제 배포가 끝나지 않으므로, GitHub Release 자동화 경로를 먼저 고정해야 첫 공식 릴리스까지 진행할 수 있기 때문

영향:

- 이제 `agentab-cli`는 `v0.1.0` 같은 태그를 push하면 GitHub Actions에서 release artifact를 만들 수 있는 상태가 되었다.
- CLI 트랙의 다음 1순위는 배포 자동화 구현이 아니라 릴리스 후보 품질 검증으로 이동했다.

후속 작업:

- `v0.1.0` 릴리스 체크리스트 실제 실행
- 릴리스 노트 초안 작성
- 첫 태그 릴리스 검증

### 2026-03-17 08:23 UTC

변경:

- `v0.1.0` 태그 릴리스 시도 중 GoReleaser가 `cmd/agentab`를 찾지 못하는 문제를 확인했다.
- 원인은 `.gitignore`의 `agentab` 패턴이 `cmd/agentab` 디렉터리까지 함께 무시하던 것이었고, 이를 `/agentab`로 수정했다.
- 빠져 있던 `cmd/agentab/main.go`를 git에 추가한 뒤 `main`에 반영했다.
- 비파괴적으로 `v0.1.1` 태그를 발행했고 GitHub Actions release가 성공했다.

이유:

- 이미 push된 `v0.1.0` 태그를 덮어쓰지 않고, 실패 원인을 수정한 새 커밋 기준으로 정상 릴리스를 만드는 것이 안전했기 때문

영향:

- `v0.1.1` GitHub Release가 실제로 생성되었고 Linux/macOS/Windows artifact와 `checksums.txt`가 업로드되었다.
- CLI 릴리스 자동화 경로가 실환경에서 한 번 검증되었다.
- 이후 릴리스부터는 같은 workflow를 반복 사용할 수 있다.

후속 작업:

- 릴리스 노트 정리
- `v0.1.0` 실패 기록을 어떻게 처리할지 결정
- CLI release checklist 항목 실제 체크 상태 반영

### 2026-03-17 08:28 UTC

변경:

- `v0.1.1` 릴리스 노트 문서와 검증 기록 문서를 추가했다.
- GitHub Release `v0.1.1` 기준으로 로컬 검증, 독립 smoke, Actions 성공 결과를 문서화했다.

이유:

- 첫 성공 릴리스의 결과와 검증 근거를 남겨야 이후 릴리스 기준선과 비교할 수 있기 때문

영향:

- 릴리스 결과가 GitHub Release 본문과 저장소 문서 양쪽에서 추적 가능해진다.
- CLI release checklist의 실제 통과 여부를 이후 별도 체크할 근거가 생겼다.

후속 작업:

- GitHub Release 본문에 릴리스 노트 반영
- CLI release checklist 항목 실제 체크 상태 반영
- `v0.1.0` 실패 태그 처리 여부 결정

### 2026-03-17 08:30 UTC

변경:

- `v0.1.1` 기준으로 CLI release checklist를 실제 검증 결과에 맞춰 채웠다.
- GitHub Release에서 다시 내려받은 공개 asset 기준으로 추가 smoke를 수행했다.
- `tab list`와 `doctor.chromeBin`은 조건부 항목으로 남겼다.

이유:

- 첫 성공 릴리스를 단순히 발행하는 것만으로는 부족하고, 어떤 항목이 실제로 통과했는지와 어떤 항목이 후속 확인이 필요한지를 명확히 남겨야 하기 때문

영향:

- CLI release checklist가 더 이상 빈 템플릿이 아니라 `v0.1.1` 기준의 실제 상태 문서가 되었다.
- 다음 보완 작업이 `tab list`와 `doctor.chromeBin` 개선으로 더 분명해졌다.

후속 작업:

- `tab list` 공개 asset smoke 이슈 재현 및 원인 분석
- `doctor`에서 `CHROME_BIN` override 노출 여부 개선
- `v0.1.0` 실패 태그 처리 여부 결정

### 2026-03-17 08:39 UTC

변경:

- `tab open` 직후 `tab list`가 stale 목록을 한 번 반환할 수 있는 race를 완화했다.
- `handleTabsList`에 현재 탭이 목록에 나타날 때까지 짧게 재시도하는 로직을 추가했다.
- `tab list` 응답에 `currentTabId`를 포함하고, 현재 탭이 목록 맨 앞에 오도록 정렬했다.
- stale-first-list 회귀 테스트를 추가했다.

이유:

- 공개 release asset smoke에서 `tab open` 직후 `tab list`가 `about:blank`만 보이는 사례가 있었고, 이를 서버 레이어에서 흡수하는 편이 CLI 사용성에 더 적합했기 때문

영향:

- `tab list`가 막 연 현재 탭을 더 안정적으로 포함하게 되었다.
- 목록 응답만 봐도 현재 탭을 식별하기 쉬워졌다.

후속 작업:

- `doctor`에서 `CHROME_BIN` override 노출 여부 개선
- 이 수정 포함 새 patch release 필요 여부 결정
- `v0.1.0` 실패 태그 처리 여부 결정

### 2026-03-17 08:42 UTC

변경:

- `doctor`가 `CHROME_BIN` override를 `chromeBin`에 반영하도록 수정했다.
- `doctor` 응답에 `chromeBinFound`, `chromeBinSource`, `chromeBinError` 필드를 추가했다.
- app 레벨 회귀 테스트를 추가하고 실제 `CHROME_BIN` override로 `agentab doctor` 출력까지 확인했다.
- 설치/트러블슈팅/릴리스 체크리스트 문서를 새 출력 기준으로 갱신했다.

이유:

- 실제 브라우저 실행은 `CHROME_BIN` override를 따르는데 `doctor`는 PATH만 보고 있어 진단 결과와 런타임 동작이 어긋났기 때문

영향:

- `doctor`만 봐도 현재 어떤 Chrome 경로가 실제로 선택될지와 그 근거를 알 수 있다.
- 공개 릴리스 기준 남은 제품 이슈는 기능 결함보다 “이 수정이 포함된 새 patch release 필요” 쪽으로 줄어들었다.

후속 작업:

- `tab list`와 `doctor.chromeBin` 수정이 포함된 새 patch release 필요 여부 결정
- `v0.1.0` 실패 태그 처리 여부 결정
- `--output text` 기준 doctor 가독성 점검

### 2026-03-17 08:53 UTC

변경:

- `v0.1.2` 태그를 생성하고 GitHub Actions release workflow를 성공시켰다.
- GitHub Release `v0.1.2`에 cross-platform artifact와 checksum이 업로드되었다.
- 공개 `linux_x86_64` asset을 다시 내려받아 별도 폴더에서 `doctor -> session start -> tab open -> tab list -> tab find -> tab click -> tab text -> session stop -> daemon stop` smoke를 수행했다.
- `tab list`와 `doctor.chromeBin` 수정이 공개 릴리스에서도 반영됨을 확인했다.
- `v0.1.2` 릴리스 노트와 검증 기록 문서를 추가했다.

이유:

- `main`에서 고친 내용을 실제 사용자가 받는 공개 release asset 기준으로 검증해야 patch release를 닫을 수 있기 때문

영향:

- `v0.1.2`는 `v0.1.1` 이후 남아 있던 대표 사용자 체감 이슈 두 개를 공개 릴리스 기준으로 해소한 버전이 되었다.
- CLI release checklist는 `v0.1.2` 기준으로 `배포 가능` 상태가 되었다.

후속 작업:

- `v0.1.2` GitHub Release 본문 정리
- `v0.1.0` 실패 태그 처리 여부 결정
- `--output text` 기준 doctor 가독성 점검

### 2026-03-18 01:25 UTC

변경:

- CLI 로드맵 문서를 `v0.1.2` 공개 릴리스 이후 상태 기준으로 재정렬했다.
- 릴리스 이전 기준으로 남아 있던 `v0.1.0` 중심 우선순위를 정리하고, 현재 하드닝 항목을 새 추천 1순위로 올렸다.
- release checklist의 `chromeBin` 중복 항목을 정리했다.

이유:

- 공개 릴리스까지 끝난 뒤에도 문서가 이전 단계 기준으로 남아 있으면 다음 작업 우선순위와 실제 제품 상태가 어긋나기 때문

영향:

- 이후 CLI 작업은 “릴리스 준비”보다 “릴리스 이후 하드닝” 관점으로 더 명확하게 추적할 수 있다.

후속 작업:

- 종료 코드와 오류 경로 테스트 보강
- 완전 새 환경 기준 PinchTab 자동 설치 smoke 재검증
- `--output text` 기준 doctor 가독성 점검

### 2026-03-18 01:31 UTC

변경:

- CLI 경로 기준 `lock_conflict`, `timeout`, `upstream_error`, 대표 `not_found` 케이스의 테스트를 추가했다.
- `response.ExitCode`가 `0/2/3/4/5/6/7`을 올바르게 반환하는지 별도 테스트를 추가했다.
- release checklist의 종료 코드 항목을 모두 완료로 갱신했다.

이유:

- 공개 릴리스 이후 남은 체크리스트 빈칸 중 가장 직접적으로 닫을 수 있는 항목이 종료 코드와 대표 오류 경로 검증이었기 때문

영향:

- CLI의 오류 코드 계약이 문서뿐 아니라 테스트 기준으로도 더 단단해졌다.
- 다음 하드닝 우선순위가 신규 환경 자동 설치 smoke와 상태 복구 쪽으로 더 분명해졌다.

후속 작업:

- 완전 새 환경 기준 PinchTab 자동 설치 smoke 재검증
- daemon 재시작 후 상태 복구 검증
- `--output text` 기준 doctor 가독성 점검

### 2026-03-18 01:36 UTC

변경:

- 완전 새 환경에서 `AGENTAB_PINCHTAB_BIN` 없이 PinchTab auto-install smoke를 다시 수행했다.
- 이 과정에서 `CHROME_BIN`을 쓸 때 PinchTab config에 server port/bind/token이 빠져 기본 `9867`로 되돌아가는 integration bug를 발견하고 수정했다.
- 현재 `main` 기준으로 `${AGENTAB_HOME}/bin/pinchtab` 설치, `session start`, `tab open`, `tab text`, `doctor` 확인까지 통과했다.

이유:

- 신규 사용자 환경에서 `CHROME_BIN`을 지정하면 auto-install이 실패할 수 있었고, 이는 공개 CLI 제품의 실제 첫 실행 경험에 직접 영향을 주는 문제였기 때문

영향:

- `main`에서는 PinchTab auto-install + `CHROME_BIN` 조합이 정상 동작한다.
- 이 수정은 아직 공개 릴리스에는 없으므로 다음 patch release 후보에 포함해야 한다.

후속 작업:

- daemon 재시작 후 상태 복구 검증
- auto-install fix 포함 next patch release 준비
- `--output text` 기준 doctor 가독성 점검

### 2026-03-18 01:41 UTC

변경:

- daemon shutdown 시 daemon metadata뿐 아니라 pinchtab metadata와 session state도 같이 정리하도록 수정했다.
- store에 `ClearSessions()`를 추가하고, server shutdown 테스트로 상태 정리 동작을 고정했다.

이유:

- daemon을 내린 뒤에도 이전 session/current tab이 남아 있으면 재시작 후 stale 상태가 남아 CLI 사용 경험이 꼬일 수 있기 때문

영향:

- `main`에서는 daemon shutdown 이후 상태가 더 예측 가능하게 정리된다.
- 신규 환경 auto-install fix와 함께 다음 patch release 가치가 더 커졌다.

후속 작업:

- auto-install / daemon cleanup fix 포함 next patch release 준비
- headed/headless 이중 smoke
- `--output text` 기준 doctor 가독성 점검

### 2026-03-18 02:06 UTC

변경:

- `v0.1.3` 태그를 생성하고 GitHub Actions release workflow를 성공시켰다.
- 공개 `linux_x86_64` asset 기준으로 auto-install, daemon cleanup, 같은 `AGENTAB_HOME`에서의 재기동까지 재검증했다.
- `v0.1.3` 릴리스 노트와 검증 기록을 추가했다.

이유:

- `main`에서 고친 auto-install / daemon state cleanup fix를 실제 사용자가 받는 공개 release asset 기준으로 검증해야 patch release를 닫을 수 있기 때문

영향:

- `v0.1.3`는 신규 환경 auto-install과 daemon shutdown state cleanup을 공개 릴리스 기준으로 보완한 버전이 되었다.
- CLI 트랙의 다음 우선순위가 더 이상 “patch release 준비”가 아니라 런타임/출력 polish로 넘어갔다.

후속 작업:

- headed/headless 이중 smoke
- `--output text` 기준 doctor 가독성 점검
- `v0.1.0` 실패 태그 처리 여부 결정

### 2026-03-18 02:12 UTC

변경:

- `scripts/smoke-modes.sh`를 추가해 headless와 headed smoke를 한 번에 재현할 수 있게 했다.
- 가상 디스플레이가 없는 머신에서는 `Xvfb`를 띄워 headed smoke를 같은 DISPLAY에서 끝까지 실행하도록 구성했다.
- 공개 `v0.1.3` asset 기준으로 `headless`, `headed` 둘 다 `session start -> tab open -> text -> find -> click -> text -> daemon stop` 흐름을 통과했다.

이유:

- 앞선 headed 시도는 각 명령을 별도 `xvfb-run`으로 감싸는 바람에 DISPLAY lifetime이 끊겨 실패했고, 이를 재사용 가능한 올바른 smoke 흐름으로 정리할 필요가 있었기 때문

영향:

- 모드별 smoke가 릴리스 이후 수동 지식이 아니라 저장소 안의 재현 가능한 스크립트가 되었다.
- CLI release checklist의 headed/headless 확인 항목을 완료로 바꿀 수 있게 되었다.

후속 작업:

- `doctor --output text` 기준 가독성 점검
- 명시적 잘못된 session/tab ID 케이스 보강
- `v0.1.0` 실패 태그 처리 여부 결정

### 2026-03-18 02:20 UTC

변경:

- `doctor` 응답을 전용 report 타입으로 바꿔 JSON 계약은 유지하면서 `--output text`에서 사람이 읽기 쉬운 섹션형 출력이 나오도록 정리했다.
- `response` 레이어에 text renderer 훅을 추가해 `doctor` 같은 명령이 pretty JSON 대신 맞춤형 text 출력을 제공할 수 있게 했다.
- app/response 테스트를 보강하고 실제 `agentab --output text doctor` 실행 결과를 설치/트러블슈팅 문서에 반영했다.

이유:

- 기존 `--output text`는 사실상 pretty JSON이라 사람이 직접 환경 문제를 읽고 판단하기에 정보 우선순위가 잘 드러나지 않았기 때문

영향:

- `doctor`는 이제 `home / chrome / pinchtab / daemon` 기준으로 핵심 상태를 바로 읽을 수 있다.
- JSON-first 계약은 그대로라 에이전트나 스크립트 소비자와의 호환성은 유지된다.

후속 작업:

- 입력/스크롤 등 액션 안정성 심화 검증
- 운영성 문서와 실제 릴리스 검증 이력 정리
- 후속 patch release 범위 결정

### 2026-03-18 02:24 UTC

변경:

- 명시적 잘못된 세션 이름과 탭 ID를 넘겼을 때 CLI가 `not_found`와 종료 코드 `4`를 일관되게 반환하는 app 레벨 테스트를 추가했다.
- `session stop`, `session resume`, `tab list`, `tab text`의 대표 경로에서 daemon이 돌려준 `not_found` 메시지가 그대로 사용자에게 전달되는지도 함께 고정했다.
- release checklist의 상태 복구 섹션에서 “잘못된 session / tab 지정 시 오류 일관성” 항목을 완료로 갱신했다.

이유:

- 기존에는 “현재 세션 없음”, “현재 탭 없음”만 닫혀 있었고, 사용자가 명시적으로 틀린 이름이나 ID를 넣는 더 현실적인 실패 케이스는 테스트 기준선이 비어 있었기 때문

영향:

- 세션명/탭 ID miss에 대한 CLI 오류 계약이 문서뿐 아니라 테스트 기준으로도 더 명확해졌다.
- 다음 1순위는 다시 로그/artifact 운영 polish 같은 사용자 경험 정리 쪽으로 넘어갈 수 있게 됐다.

후속 작업:

- 로그 / artifact 운영 polish
- `v0.1.0` 실패 태그 처리 여부 결정
- 입력/스크롤 등 액션 안정성 심화 검증

### 2026-03-18 02:33 UTC

변경:

- `doctor` 응답에 `logsDir`, `daemonLogPath`, `pinchtabLogPath`를 추가하고 text 출력에서도 바로 보이도록 했다.
- 관리형 artifact 저장 응답에 `managed`, `relativePath`, `createdAt` 메타데이터를 추가했다.
- store/app 테스트를 보강하고 설치/트러블슈팅/스펙/체크리스트 문서를 현재 운영 정보에 맞게 갱신했다.

이유:

- 사용자가 장애 상황에서 로그 파일과 artifact 위치를 다시 찾기 위해 문서와 디렉터리를 뒤지는 시간을 줄이고, 저장 응답만 봐도 관리형 artifact인지와 공유 가능한 상대 경로를 바로 알 수 있게 하려는 목적

영향:

- `doctor`만으로 현재 홈, 로그 위치, artifact 루트, Chrome 경로, PinchTab 상태를 한 번에 확인할 수 있다.
- artifact 저장 결과가 후속 디버깅, 리포트, 이슈 공유에 더 직접적으로 쓰일 수 있게 됐다.

후속 작업:

- 입력/스크롤 등 액션 안정성 심화 검증
- 운영성 문서와 실제 릴리스 검증 이력 정리
- 후속 patch release 범위 결정

### 2026-03-18 02:49 UTC

변경:

- 실패했던 `v0.1.0` 태그를 로컬과 GitHub 원격에서 삭제했다.
- 현재 기준 문서에서 `v0.1.0`이 여전히 남아 있는 것처럼 보이던 표현을 정리하고, `v0.1.1`을 실제 첫 공개 릴리스로 명확히 적었다.

이유:

- `v0.1.0`은 GitHub Release도 없는 실패 시도 흔적이었고, 태그만 남아 있으면 외부 사용자에게 버전 히스토리를 불필요하게 헷갈리게 만들 수 있었기 때문

영향:

- GitHub에서 보는 공개 버전 히스토리가 `v0.1.1`부터 시작하는 실제 사용 흐름과 더 잘 맞게 됐다.
- 앞으로는 release note와 체크리스트 기준도 “실제 첫 공개 릴리스는 `v0.1.1`”로 일관되게 설명할 수 있다.

후속 작업:

- 입력/스크롤 등 액션 안정성 심화 검증
- 운영성 문서와 실제 릴리스 검증 이력 정리
- 후속 patch release 범위 결정

### 2026-03-18 03:01 UTC

변경:

- `testdata/smoke/index.html`을 확장해 `click`, `type`, `fill`, `press`, `scroll` 상태 변화를 한 페이지에서 검증할 수 있게 했다.
- `scripts/smoke-modes.sh`를 확장해 `headless`와 `headed` 모두에서 액션 smoke를 수행하고 각 액션의 성공 여부를 요약 출력하도록 만들었다.
- PinchTab local runtime이 새 홈마다 기본 `9867` 대신 사용 가능한 loopback 포트를 고르도록 하드닝해 smoke 재현 안정성을 높였다.

이유:

- 체크리스트에서 남아 있던 액션 안정성 항목을 실제 브라우저 기준으로 닫으려면 `click`뿐 아니라 `type`, `fill`, `press`, `scroll`까지 재현 가능한 smoke 흐름이 필요했고, 고정 포트 의존성은 반복 검증을 흔들 수 있었기 때문

영향:

- `scripts/smoke-modes.sh` 기준으로 `headless`, `headed` 둘 다 `click_ok=True`, `type_ok=True`, `fill_ok=True`, `press_ok=True`, `scroll_ok=True`를 확인할 수 있게 됐다.
- release checklist의 액션 안정성 항목을 완료로 전환할 수 있게 됐다.
- `${AGENTAB_HOME}`별 PinchTab runtime 분리가 더 안정적으로 동작한다.

후속 작업:

- 운영성 문서와 실제 릴리스 검증 이력 정리
- 후속 patch release 범위 결정
- 필요하면 공개 release asset 기준 액션 smoke 한 번 더 수행

### 2026-03-18 04:36 UTC

변경:

- 운영 관점에서 바로 참고할 [cli-operations-runbook.md](/workspace/agentab-cli/docs/cli-operations-runbook.md)를 추가했다.
- 공개 릴리스별 note/verification 문서를 한눈에 보는 [releases/README.md](/workspace/agentab-cli/docs/releases/README.md) 인덱스를 추가했다.
- README, CLI 개요, release workflow 문서에서 새 운영/검증 문서로 연결되는 링크를 보강했다.

이유:

- 실제 운영 중에는 설치 문서, 트러블슈팅 문서, 릴리스 검증 문서가 흩어져 있으면 필요한 정보를 다시 찾는 데 시간이 오래 걸리고, 현재 공개 릴리스 기준선과 `main` 하드닝 상태도 한눈에 비교하기 어려웠기 때문

영향:

- 운영자는 runbook 한 장으로 `doctor -> logs -> artifacts -> smoke` 흐름을 바로 따라갈 수 있다.
- 릴리스별 공개 검증 기준선과 아직 릴리스에 올라가지 않은 `main` 하드닝 항목을 구분해서 설명할 수 있게 됐다.

후속 작업:

- 후속 patch release 범위 결정
- 액션 smoke를 다음 patch release 검증 기록에 반영
- 공개 release asset 기준 액션 smoke 승격 여부 결정
