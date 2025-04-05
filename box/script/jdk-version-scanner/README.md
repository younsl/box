# JDK 버전 스캐너

## 개요

모든 파드의 JDK 버전을 스캔하여 출력합니다.

## 사용 방법

`maxGoroutines` 상수를 설정하여 최대 동시 실행 고루틴 개수를 설정할 수 있습니다.

```go
const maxGoroutines = 20
```

`go run` 명령어를 사용하여 실행합니다.

```bash
go run -v main.go
```

## 출력 결과

출력 결과는 다음과 같습니다.

```bash
Index  Namespace   Pod                             Java Version
1      default     example-pod-67c49d5ff8-sxm2z    21.0.2
2      default     example-pod-6974df45d-sv4cq     17.0.4.1
...
97     default  example-pod-7b64fc9986-ncmh6    17.0.9
Total pods scanned: 343
Pods using JDK: 97
Time taken: 2m 34s
```
