## 개요

configMap을 변경해도 해당 configMap을 마운트하고 있는 파드가 재시작되지 않는 이슈가 있었습니다.

이를 항상 kubectl rollout restart 등으로 매번 재시작해주는 게 번거로워서 개선이 필요했습니다.

## 개선 방법

configMap에 들어있는 데이터의 내용이 변경되면 재시작할 수 있게, `checksum/config` annotation을 파드에 부여합니다.

stakater의 reloader를 사용하지 않은 이유는 서드파티 관리 컨트롤러를 하나 더 늘리고 싶지 않고 쿠버네티스 네이티브한 기능만 가지고 해결하기 위함이었습니다.

## 적용 방법

개선하려고 하는 헬름 차트 구조는 다음과 같습니다.

```console
$ tree -L 2 .
.
├── Chart.yaml
├── installation.md
├── templates
│   ├── NOTES.txt
│   ├── _helpers.tpl
│   ├── configmap.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── values.yaml
 ```

`_helpers.tpl` 파일에 다음과 같이 configMap checksum 값을 계산하는 템플릿을 추가합니다.


```go
{{/*
Generate a hash of the configmap to trigger pod restarts
if the configMap data is changed.
*/}}
{{- define "kafka-connect-ui.configmap.checksum" }}
checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
{{- end }}
```

`_helpers.tpl`의 kafka-connect-ui.configmap.checksum 템플릿에서 계산된 checksum 값을 pod annotation에 넣도록 `deployment.yaml`을 변경합니다.

```
spec:
  template:
    metadata:
      annotations:
        {{- include "kafka-connect-ui.configmap.checksum" . | nindent 8 }}
```

실제 헬름 차트로 배포된 파드에는 `checksum/config` annotation이 추가됩니다.

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    checksum/config: 7d26f9f86ba6c4413b41ddfe08e125d80b50fb6ed23aaef9be557af2dd972a32
    kubectl.kubernetes.io/restartedAt: "2024-08-07T16:52:13+09:00"
```

configMap 데이터가 업데이트될 때마다 `checksum/config` annotation 값도 같이 바뀌므로, 결과적으로 파드가 새 configMap의 데이터를 자동으로 들고오게 됩니다.

## 관련자료

[Leveraging Helm for ConfigMap Updates](https://www.baeldung.com/ops/kubernetes-restart-configmap-updates#3-leveraging-helm-for-configmap-updates)
