# agentab-cli 릴리스 검증 이력

상태: 초안  
작성일: 2026-03-18  
마지막 갱신: 2026-03-18 03:01 UTC
목적: 공개 릴리스별로 무엇이 검증되었는지와 어떤 문서를 보면 되는지 한눈에 정리하기 위함

## 공개 릴리스 기준선

실제 첫 공개 릴리스는 `v0.1.1`이다.

| 버전 | 릴리스 노트 | 검증 기록 | 핵심 의미 |
| --- | --- | --- | --- |
| `v0.1.1` | [v0.1.1.md](/workspace/agentab-cli/docs/releases/v0.1.1.md) | [v0.1.1-verification.md](/workspace/agentab-cli/docs/releases/v0.1.1-verification.md) | 첫 성공 공개 릴리스 |
| `v0.1.2` | [v0.1.2.md](/workspace/agentab-cli/docs/releases/v0.1.2.md) | [v0.1.2-verification.md](/workspace/agentab-cli/docs/releases/v0.1.2-verification.md) | `tab list`와 `doctor.chromeBin` 보강 |
| `v0.1.3` | [v0.1.3.md](/workspace/agentab-cli/docs/releases/v0.1.3.md) | [v0.1.3-verification.md](/workspace/agentab-cli/docs/releases/v0.1.3-verification.md) | auto-install / daemon cleanup / headed-headless smoke 보강 |
| `v0.1.4` | [v0.1.4.md](/workspace/agentab-cli/docs/releases/v0.1.4.md) | [v0.1.4-verification.md](/workspace/agentab-cli/docs/releases/v0.1.4-verification.md) | 운영성 개선 / action smoke 확대 / runtime 포트 하드닝 |

## 현재 `main`의 추가 하드닝

아래 항목은 `main`에서 검증됐지만 아직 공개 release verification 문서에 승격되지 않은 항목이다.

- `doctor --output text` 운영 정보 정리
- 로그 / artifact 진단 메타데이터 보강
- 명시적 잘못된 session / tab ID 오류 경로 테스트 보강
- `click`, `type`, `fill`, `press`, `scroll` action smoke 확대
- PinchTab local runtime 포트 선택 하드닝

다음 patch release를 만들면, 해당 버전의 verification 문서에 이 항목들을 승격해 기록한다.

## 운영 문서

운영 관점에서 바로 참고할 문서:

- [operations runbook](/workspace/agentab-cli/docs/cli-operations-runbook.md)
- [release workflow](/workspace/agentab-cli/docs/cli-release-workflow.md)
- [release checklist](/workspace/agentab-cli/docs/cli-release-checklist.md)
