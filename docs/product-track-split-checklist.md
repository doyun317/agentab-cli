# agentab 제품 트랙 분리 체크리스트

상태: 작업 기준선  
작성일: 2026-03-17  
목적: `agentab CLI` 제품 트랙과 `agentab-langchain` 통합 트랙을 명확히 분리해, 앞으로의 작업과 출시 기준을 흔들림 없이 관리하기 위함

## 1. 한 줄 정리

- 제품 1: `agentab`
  에이전트가 브라우저를 조작하기 위해 사용하는 로컬 우선 CLI
- 제품 2: `agentab-langchain`
  LangChain이 `agentab` CLI를 쉽게 사용할 수 있게 해주는 공식 adapter

핵심 원칙:

- `agentab`은 본체다.
- `agentab-langchain`은 공식 통합체다.
- planner, benchmark, prompt tuning은 기본적으로 adapter 쪽에 둔다.
- 브라우저 lifecycle, session/tab, JSON 계약, 안정성은 기본적으로 CLI 쪽에 둔다.

## 2. 경계 원칙

### 2.1 `agentab CLI`에 들어가야 하는 것

- 브라우저 설치 확인과 부트스트랩
- PinchTab 기동과 생명주기 관리
- daemon 자동 기동과 상태 확인
- session / tab / lock 관리
- `open`, `list`, `snapshot`, `text`, `find`, `click`, `type`, `press`, `scroll`, `screenshot` 같은 stable primitive
- JSON envelope와 종료 코드
- timeout, retry, 오류 표준화
- trace, artifact, report를 저장하기 위한 low-level 기반

### 2.2 `agentab-langchain`에 들어가야 하는 것

- LangChain tool wrapper
- model preset
- prompt template
- planner / executor / verifier workflow
- LangChain-specific recovery layer
- benchmark harness
- trace manifest를 해석하고 비교하는 상위 로직

### 2.3 넣지 말아야 하는 것

- CLI 본체에 LangChain 전용 개념을 직접 넣지 않는다.
- adapter 쪽에서 브라우저 lifecycle을 다시 구현하지 않는다.
- CLI에 특정 모델용 prompt workaround를 넣지 않는다.
- adapter에서 PinchTab 내부 세부사항을 직접 알도록 만들지 않는다.

## 3. 의사결정 질문

새 작업이 들어오면 먼저 아래 질문으로 분류한다.

1. 이 변경이 LangChain 없이도 모든 에이전트에게 유용한가?
   그렇다면 `agentab CLI` 트랙이다.
2. 이 변경이 LangChain agent loop, prompt, planner 품질을 위한 것인가?
   그렇다면 `agentab-langchain` 트랙이다.
3. 이 변경이 둘 다에 영향을 주는가?
   먼저 CLI 계약을 정리하고, 그 다음 adapter 적용 작업을 분리한다.

## 4. 트랙 A: `agentab CLI` 제품 체크리스트

목표:

- 누구나 설치해서 바로 사용할 수 있는 agent-friendly browser control CLI를 배포 가능한 상태로 만든다.

### 4.1 제품 정의

- [ ] 제품 이름과 설명 문구 고정
- [ ] 지원 플랫폼 범위 확정
- [ ] 지원 브라우저/런타임 정책 확정
- [ ] 버전 정책 정리

### 4.2 핵심 기능

- [ ] `doctor` 명령 완성도 점검
- [ ] `daemon start|status|stop` 안정화
- [ ] `session start|list|stop` 안정화
- [ ] `tab open|list|close` 안정화
- [ ] `snapshot`, `text`, `find`, `click`, `type`, `press`, `scroll`, `screenshot` 계약 점검
- [ ] 모든 명령의 JSON 출력 형식 일관성 점검
- [ ] 종료 코드 표준화 문서화

### 4.3 신뢰성

- [ ] auto-start daemon 안정성 점검
- [ ] PinchTab 설치/업데이트 경로 안정화
- [ ] Chrome 런타임 확인 로직 안정화
- [ ] timeout / upstream error / lock conflict 처리 점검
- [ ] 상태 저장과 재시작 복구 점검

### 4.4 운영성

- [ ] 기본 trace / artifact 저장 경로 확정
- [ ] screenshot / snapshot artifact 저장 구조 완성
- [ ] manifest / report 연결 구조 점검
- [ ] 로그와 디버그 출력 기준 정리

### 4.5 문서와 배포

- [ ] 설치 문서 정리
- [ ] CLI 사용 문서 정리
- [ ] 문제 해결 문서 정리
- [ ] release checklist 작성
- [ ] 바이너리 배포 방식 결정

### 4.6 출시 완료 기준

- [ ] 새 머신에서 `doctor -> session start -> tab open -> text -> click -> screenshot -> session stop`이 문서대로 재현된다.
- [ ] core 테스트와 fake PinchTab 통합 테스트가 안정적으로 통과한다.
- [ ] 실제 live smoke가 최소 기준 모델/페이지에서 재현된다.
- [ ] CLI만으로도 사람이 직접 디버깅 가능한 수준의 문서와 오류 메시지가 있다.

## 5. 트랙 B: `agentab-langchain` 통합 체크리스트

목표:

- LangChain agent가 `agentab` CLI를 편하게 쓰도록 공식 adapter와 reference workflow를 제공한다.

### 5.1 제품 정의

- [ ] 패키지 이름과 공개 범위 확정
- [ ] 최소 지원 LangChain 버전 확정
- [ ] 공식 지원 모델 preset 범위 확정
- [ ] adapter의 비목표 명시

### 5.2 기본 통합

- [ ] `agentab` CLI subprocess client 안정화
- [ ] tool schema와 설명 다듬기
- [ ] schema normalization / recovery layer 유지보수
- [ ] example agent 흐름 정리
- [ ] planner / executor / verifier helper 정리

### 5.3 품질

- [ ] single-agent benchmark 유지
- [ ] planner-executor benchmark 유지
- [ ] 실제 로컬 모델 preset 검증
- [ ] 실패 유형 분류 기준 정리
- [ ] inspection loop 감지 규칙 추가

### 5.4 디버깅과 평가

- [ ] workflow benchmark 결과 누적
- [ ] trace manifest를 활용한 비교 도구 추가
- [ ] screenshot / snapshot artifact를 adapter 분석 흐름에 연결
- [ ] regression 비교 리포트 정리

### 5.5 문서

- [ ] LangChain quickstart 정리
- [ ] local model preset 문서화
- [ ] planner-executor 예제 문서화
- [ ] benchmark 실행 가이드 정리

### 5.6 출시 완료 기준

- [ ] LangChain 사용자가 `agentab` 설치 후 예제 agent를 바로 실행할 수 있다.
- [ ] 두 개 이상의 실제 모델 preset에서 benchmark가 재현된다.
- [ ] 실패 시 trace / manifest / report만으로 원인 분석이 가능하다.
- [ ] adapter가 CLI 본체의 계약을 깨지 않고 독립적으로 개선될 수 있다.

## 6. 작업 순서 권장안

권장 순서:

1. `agentab CLI`를 먼저 제품처럼 마무리한다.
2. CLI 계약이 안정되면 `agentab-langchain`을 공식 adapter로 정리한다.
3. planner / benchmark / prompt 최적화는 adapter 트랙에서 빠르게 실험한다.
4. 실험 결과 중 범용적인 것은 다시 CLI 본체에 반영한다.

핵심:

- CLI를 먼저 고정한다.
- adapter는 그 위에서 빠르게 실험한다.
- 실험과 본체를 섞지 않는다.

## 7. 지금 기준 권장 우선순위

### 7.1 `agentab CLI`

- [ ] screenshot / snapshot artifact 저장 구조 완성
- [ ] CLI release checklist 초안 작성
- [ ] 설치/실행/트러블슈팅 문서 보강

### 7.2 `agentab-langchain`

- [ ] trace manifest를 활용한 비교/분석 흐름 추가
- [ ] inspection loop 감지 규칙 추가
- [ ] planner-executor benchmark를 더 체계화

## 8. 한 문장 기준선

앞으로 작업 중 헷갈리면 이 문장을 기준으로 판단한다.

“`agentab`은 배포 가능한 브라우저 조작 CLI이고, `agentab-langchain`은 그 CLI를 LangChain에서 활용하기 위한 공식 adapter다.”
