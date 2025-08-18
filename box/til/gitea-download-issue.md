# Gitea 대용량 파일 다운로드 이슈 분석

## 증상

### 주요 현상
사용자가 Gitea 릴리즈 페이지에서 `.tar.gz` 파일 다운로드 시 간헐적인 실패 발생
- 여러 번 재시도 후 최종적으로 다운로드 성공
- 주로 대용량 파일에서 발생
- 브라우저에서 다운로드가 중간에 끊김

### 환경 정보
- **Gitea**: v1.24.2-rootless
- **Kubernetes**: EKS 1.32
- **Cache**: Valkey (Redis 호환)
- **인증 방식**: 일반 ID/Password

### 상세 증상 분석

#### 1. HTTP Range 요청 패턴 (tar.gz 다운로드)
- 브라우저가 대용량 파일을 청크 단위로 나누어 다운로드 시도
- 첫 번째 청크는 성공 (206 Partial Content, ~190ms 소요)
- **즉시** 다음 청크 요청 시 인증 실패 (401 Unauthorized, ~7ms만에 실패)
- 핵심: 206 응답 직후 10-30ms 이내에 context canceled 발생

#### 2. 시간대별 발생 패턴
```
07:59:01 - tar.gz 다운로드 실패 (206 Partial Content → 401 Unauthorized)
07:59:12 - tar.gz 재시도 실패 (206 Partial Content → 401 Unauthorized) 
08:01:17 - tar.gz 다운로드 실패 (206 Partial Content → 401 Unauthorized)
08:02:43 - ZIP 아카이브 실패 (500 Internal Server Error) ← 별개 이슈
08:03:06 - tar.gz 다운로드 성공 (200 OK, 1051ms)
08:03:47~08:04:33 - 캐시 응답 (304 Not Modified)
```

#### 3. 401 Unauthorized 에러 발생 시퀀스 
```
1. GET 요청 → 206 Partial Content (성공)
2. 10ms 이내 → GetUserByID: context canceled 
3. 세션 검증 실패 → Failed to verify user
4. 401 Unauthorized 응답 (총 ~7ms)
```

### 발생 URL 예시
```
http://gitea.example.com/OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz
```

## 로그 분석

### 주요 에러 패턴
```
2025-08-18T07:59:01.290779720Z gitea 2025/08/18 07:59:01 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.1:18329, 206 Partial Content in 192.5ms @ repo/repo.go:318(repo.RedirectDownload)
2025-08-18T07:59:01.320181025Z gitea 2025/08/18 07:59:01 services/auth/session.go:51:(*Session).Verify() [E] GetUserByID: context canceled
2025-08-18T07:59:01.320228848Z gitea 2025/08/18 07:59:01 routers/web/web.go:121:Routes.webAuth.10() [E] Failed to verify user: context canceled
2025-08-18T07:59:01.322235910Z gitea 2025/08/18 07:59:01 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.1:18329, 401 Unauthorized in 6.8ms @ web/web.go:118(web.Routes.webAuth)
2025-08-18T07:59:12.250175593Z gitea 2025/08/18 07:59:12 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.2:57323, 206 Partial Content in 193.7ms @ repo/repo.go:318(repo.RedirectDownload)
2025-08-18T07:59:12.262348544Z gitea 2025/08/18 07:59:12 services/auth/session.go:51:(*Session).Verify() [E] GetUserByID: context canceled
2025-08-18T07:59:12.262387879Z gitea 2025/08/18 07:59:12 routers/web/web.go:121:Routes.webAuth.10() [E] Failed to verify user: context canceled
2025-08-18T07:59:12.264366790Z gitea 2025/08/18 07:59:12 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.2:57323, 401 Unauthorized in 6.8ms @ web/web.go:118(web.Routes.webAuth)
```

### 유사 패턴 반복 (08:01)
```
2025-08-18T08:01:17.477449999Z gitea 2025/08/18 08:01:17 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.3:3633, 206 Partial Content in 189.9ms @ repo/repo.go:318(repo.RedirectDownload)
2025-08-18T08:01:17.488702742Z gitea 2025/08/18 08:01:17 services/auth/session.go:51:(*Session).Verify() [E] GetUserByID: context canceled
2025-08-18T08:01:17.488750103Z gitea 2025/08/18 08:01:17 routers/web/web.go:121:Routes.webAuth.10() [E] Failed to verify user: context canceled
2025-08-18T08:01:17.491089433Z gitea 2025/08/18 08:01:17 HTTPRequest [I] router: completed GET /OrgName/repo_name/releases/download/v1.0.35/build-artifacts-v1.0.35.tar.gz for 10.0.0.3:3633, 401 Unauthorized in 6.6ms @ web/web.go:118(web.Routes.webAuth)
```

### 응답 코드 시퀀스
1. **206 Partial Content**: 첫 번째 청크 다운로드 성공
2. **401 Unauthorized**: 즉시(~10ms 이내) 다음 청크 요청 시 인증 실패
3. **200 OK / 304 Not Modified**: 재시도 후 성공

## 예상 원인

### 1. HTTP Range 요청 처리 이슈 (핵심 문제)
**근거:**
- 206 Partial Content 응답 직후 10ms 이내에 `context canceled` 발생
- 동일한 파일, 동일한 사용자가 즉시 재요청 시 401 발생
- Range 요청이 아닌 일반 다운로드(200 OK)는 성공

**메커니즘:**
- 브라우저가 대용량 파일을 Range 헤더로 청크 단위 요청
- 첫 청크 응답(206) 완료 시 HTTP 연결 컨텍스트가 종료
- 다음 청크 요청 시 세션 검증(`GetUserByID`)이 이미 취소된 컨텍스트 사용
- Gitea 1.24.2-rootless의 세션 미들웨어가 컨텍스트 생명주기를 제대로 관리하지 못함

### 2. 짧은 타임아웃 설정
- Gitea 기본 타임아웃: 60초 (`TIMEOUT_READ`, `TIMEOUT_WRITE`)
- 하지만 에러가 10ms 이내 발생하므로 타임아웃이 직접 원인은 아님
- 다만 대용량 파일 전체 다운로드 시간에는 영향

### 3. Valkey(Redis) 세션 스토어 관련
- `PROVIDER_CONFIG`의 `idle_timeout=180s`는 충분
- 하지만 컨텍스트가 취소되면 Redis 조회 자체가 실패
- rootless 이미지 특성상 세션 처리 경로가 다를 수 있음

## 현재 설정

### app.ini 주요 설정

```ini
[server]
PROTOCOL = http
HTTP_PORT = 443
DOMAIN = gitea.example.com

[session]
PROVIDER = redis
PROVIDER_CONFIG = redis+cluster://:@redis-cluster-headless.gitea.svc.cluster.local:6379/0?pool_size=100&idle_timeout=180s&

[database]
DB_TYPE = postgres
HOST = postgresql-ha-pgpool.gitea.svc.cluster.local:5432
```

**문제점**: 타임아웃 관련 설정이 누락되어 기본값(60초) 사용 중

## 해결 방안

### 1. Gitea 설정 개선 (Helm values.yaml)

```yaml
gitea:
  config:
    server:
      # 타임아웃 설정 증가 (30분)
      TIMEOUT_WRITE: "1800s"
      TIMEOUT_READ: "1800s"
      # 대용량 파일 압축 비활성화
      ENABLE_GZIP: false
      
    session:
      # 세션 유지 시간 연장 (24시간)
      SESSION_LIFE_TIME: 86400
      COOKIE_NAME: "i_like_gitea"
      GC_INTERVAL_TIME: 86400
      # Redis 연결 타임아웃 증가
      PROVIDER_CONFIG: "redis+cluster://:@redis-cluster-headless.gitea.svc.cluster.local:6379/0?pool_size=100&idle_timeout=300s&read_timeout=30s&write_timeout=30s"
      
    repository:
      upload:
        # 최대 파일 크기 증가
        FILE_MAX_SIZE: 500  # MB
        MAX_FILES: 10
        
    web:
      # 대용량 파일 로깅 부하 감소
      ENABLE_ACCESS_LOG: false
```

### 2. 배포 명령

```bash
helm upgrade gitea ./gitea -f values.yaml
```

### 3. 클라이언트 측 우회 방법

다운로드 실패 시 아래 명령 사용:
```bash
# wget으로 재시도 자동화
wget --continue <URL>

# curl로 단일 연결 다운로드
curl -L -O <URL>

# 부분 다운로드 재개
curl -C - -O <URL>
```

## 추가 고려사항

1. **리버스 프록시/로드밸런서 타임아웃**
   - Ingress Controller나 외부 프록시의 타임아웃도 함께 확인 필요

2. **Gitea 버전 업그레이드**
   - 최신 버전에서 Range 요청 처리 개선 여부 확인

3. **모니터링**
   - 다운로드 실패율 메트릭 수집
   - 세션 만료 빈도 추적

## 참고 자료

- [Gitea Configuration Cheat Sheet](https://docs.gitea.io/en-us/config-cheat-sheet/)
- 인증 방식: 일반 ID/Password 기반 (OAuth 미사용)
- 환경: Kubernetes 클러스터 내 Helm 차트로 배포
