# agentab CLI 로드맵 및 작업 기록

상태: CLI 본체 전용 운영 문서  
최초 작성: 2026-03-17  
마지막 갱신: 2026-03-17 06:19 UTC
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

- CLI 제품 기준 smoke 시나리오 정리
- 운영 로그와 artifact 기준 정리
- 배포 방식 결정

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
- `todo` 명령별 text output 경험 점검
- `todo` CLI 예제 명령 셋 정리

### 5.2 런타임과 신뢰성

- `done` PinchTab 자동 설치 경로 구현
- `done` daemon auto-start 경로 구현
- `done` PinchTab / daemon child process detach 보완
- `done` 최신 PinchTab의 `browser.binary` 전달 보완
- `done` 멀티 인스턴스 `tab list` 경로 수정
- `todo` Chrome 런타임 의존성 점검 흐름 개선
- `todo` lock / timeout / upstream error 문서화

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
- `todo` git 저장소 생성 후 GitHub Actions release workflow 연결

## 6. 현재 추천 1순위

- git 저장소 생성 후 GitHub Actions release workflow 연결

이유:

- 배포 방식은 `GitHub Releases + GoReleaser`로 결정되었고 로컬 snapshot 검증 경로도 생겼으므로, 다음 병목은 실제 저장소와 태그 기반 릴리스 자동화 연결이기 때문

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
