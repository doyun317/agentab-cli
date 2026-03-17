# agentab CLI 배포 준비 워크플로우

상태: 초안  
작성일: 2026-03-17  
마지막 갱신: 2026-03-17 08:42 UTC
목적: `agentab CLI`의 배포 방식을 `GitHub Releases + GoReleaser`로 고정하고, GitHub 저장소 생성 전후에 무엇을 해야 하는지 분리해 정리하기 위함

## 1. 선택한 배포 방식

현재 `agentab CLI`의 기본 배포 방식은 아래 조합으로 고정한다.

- 배포 채널: `GitHub Releases`
- 빌드/아카이브 자동화: `GoReleaser`

이 선택의 의미:

- 사용자는 GitHub Release에서 플랫폼별 바이너리와 체크섬을 받는다.
- 개발자는 GoReleaser 설정 하나를 기준으로 로컬 검증과 실제 릴리스를 모두 반복한다.

## 2. 지금 가능한 범위

현재 `agentab-cli`는 GitHub 저장소와 연결되어 있고 release workflow 파일도 존재한다.

현재 가능한 것은 다음이다.

- GoReleaser 설정 검증
- cross-platform snapshot binary build 검증
- GitHub Actions 기반 tag release 실행
- release 직전까지 필요한 스크립트와 문서 정리

아직 남아 있는 것은 다음이다.

- 공개 릴리스 기준 patch 변경사항 반영
- release notes 확정
- 릴리스 체크리스트 최종 통과

## 3. 로컬 사전 검증 흐름

### 3.1 도구 설치

```bash
cd /workspace/agentab-cli
./scripts/bootstrap-tools.sh
./scripts/install-goreleaser.sh
```

설치 결과:

- Go: `/workspace/agentab-cli/.tools/go`
- GoReleaser: `/workspace/agentab-cli/.tools/bin/goreleaser`

## 3.2 snapshot build 검증

```bash
cd /workspace/agentab-cli
./scripts/release-snapshot.sh
```

이 스크립트가 하는 일:

1. GoReleaser 설치 확인
2. git 저장소 여부 확인
3. git 저장소가 있으면 `goreleaser check`
4. git 저장소 유무와 관계없이 `goreleaser release --snapshot --clean`

현재 로컬에서는 snapshot release 검증에 사용하고, 원격 GitHub에서는 tag push 시 release workflow가 실행된다.

주의:

- `goreleaser check`는 git 저장소 바깥에서 실패한다.
- 하지만 `goreleaser release --snapshot --clean`은 non-git 환경에서도 동작하며 archive와 checksum까지 만든다.
- 그래서 현재 스크립트는 non-git 환경에서는 `check`를 건너뛰고 snapshot release 자체를 검증 경로로 사용한다.

산출물 위치:

- 기본 출력 디렉터리: `/workspace/agentab-cli/dist`

예시 산출물:

- `agentab-cli_0.1.2-snapshot_linux_x86_64.tar.gz`
- `agentab-cli_0.1.2-snapshot_linux_arm64.tar.gz`
- `agentab-cli_0.1.2-snapshot_macOS_x86_64.tar.gz`
- `agentab-cli_0.1.2-snapshot_macOS_arm64.tar.gz`
- `agentab-cli_0.1.2-snapshot_windows_x86_64.zip`
- `agentab-cli_0.1.2-snapshot_windows_arm64.zip`
- `checksums.txt`

참고:

- `dist/agentab_linux_amd64_v1/` 같은 디렉터리 이름은 GoReleaser 내부 build 산출물 경로다.
- 사용자가 직접 받게 되는 릴리스 asset 이름은 위의 archive 이름처럼 정리된다.

## 4. git 저장소가 생긴 뒤 해야 할 일

### 4.1 최소 조건

- `agentab-cli` 코드가 git 저장소 안에 있어야 한다.
- 실제 release를 올릴 원격 GitHub 저장소가 있어야 한다.
- `.github/workflows/release.yml`이 있어야 한다.

### 4.2 그 다음 단계

1. 로컬 git 저장소 초기화 또는 기존 저장소 연결
2. 원격 GitHub 저장소 연결
3. tag 규칙 확정
4. GitHub Actions release workflow 추가
5. patch 태그로 release 실행

현재 상태:

- `1` 완료
- `2` 완료
- `3` 진행 중
- `4` 완료
- `5` 반복 가능

현재 workflow 파일:

- `/workspace/agentab-cli/.github/workflows/release.yml`

## 5. 현재 GoReleaser 설정 범위

현재 `.goreleaser.yaml`은 아래를 기준으로 한다.

- `./cmd/agentab`에서 `agentab` 바이너리 빌드
- `linux`, `darwin`, `windows`
- `amd64`, `arm64`
- `CGO_ENABLED=0`
- snapshot 버전은 `0.1.2-snapshot`
- changelog는 일단 비활성화

의도:

- GitHub 저장소가 생기기 전에도 설정 검증과 build 검증이 가능해야 한다.
- GitHub 저장소가 생기기 전에도 사용자가 받게 될 archive 이름을 미리 검증할 수 있어야 한다.
- tag 기반 실제 릴리스는 나중에 덧붙이되, 설정의 중심은 지금부터 고정한다.

## 6. 다음 단계

배포 준비 기준 다음 순서:

1. patch 변경사항 커밋 및 `main` 반영
2. 릴리스 노트 정리
3. `v0.1.2` 태그 푸시
4. GitHub Release 결과 검증
