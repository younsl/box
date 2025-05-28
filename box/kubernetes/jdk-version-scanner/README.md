# JDK Version Scanner

Kubernetes 클러스터에서 실행 중인 Pod들의 Java 버전을 스캔하는 도구입니다.

## 기능

- 여러 네임스페이스의 Pod 동시 스캔
- DaemonSet Pod 필터링
- 병렬 처리 (configurable goroutine 수)
- 타임아웃 설정
- Graceful shutdown
- 상세한 결과 출력

## 설치

```bash
go build -o jdk-scanner ./cmd/scanner
```

## 사용법

```bash
# 기본 사용법 (default 네임스페이스)
./jdk-scanner

# 여러 네임스페이스 스캔
./jdk-scanner -namespaces="default,kube-system,monitoring"

# 고급 옵션
./jdk-scanner \
  -namespaces="default,app" \
  -max-goroutines=30 \
  -timeout=60s \
  -skip-daemonset=false \
  -verbose
```

## 옵션

- `-namespaces`: 스캔할 네임스페이스 (쉼표로 구분)
- `-max-goroutines`: 최대 동시 실행 goroutine 수 (기본: 20)
- `-timeout`: kubectl 명령 타임아웃 (기본: 30s)
- `-skip-daemonset`: DaemonSet Pod 건너뛰기 (기본: true)
- `-verbose`: 상세 로그 출력 (기본: false)

## 요구사항

- `kubectl` 명령어가 설치되어 있어야 함
- Kubernetes 클러스터에 접근 권한이 있어야 함
- Pod에서 `java -version` 명령어 실행 가능해야 함

## 출력 예시

```
INDEX   NAMESPACE   POD                    JAVA_VERSION
1       default     app-deployment-abc123  11.0.16
2       default     api-service-def456     1.8.0_292
3       monitoring  prometheus-ghi789      17.0.2

Scan Summary:
Total pods scanned: 15
Pods using JDK: 3
Time taken: 1m 23s
```