# OpenTelemetry

- GraalVM은 일반적인 기존 Newrelic JVM Agent를 통해 메트릭을 보낼 수 없으며, OpenTelemetry SDK를 사용해서 계측한 후, OpenTelemetry Exporter를 통해 뉴렐릭으로 내보내는 방식만 지원합니다. 자세한 사항은 [Introduction to OpenTelemetry and New Relic](https://docs.newrelic.com/docs/opentelemetry/opentelemetry-introduction/)을 참고합니다.
- OTel 1.47.0 기준 [OTEL_TRACES_SAMPLER_ARG 환경변수](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration)를 통해 샘플링 비율을 0에서 1사이로 조절 가능함. 기본값 1(100%)임.

<details>
<summary>Example</summary>

OTel configuration example for kubernetes deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jvm-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: jvm-app
  template:
    metadata:
      labels:
        app: jvm-app
    spec:
      containers:
      - name: jvm-app
        image: your-registry/jvm-app:latest
        env:
        - name: OTEL_TRACES_SAMPLER
          value: "parentbased_traceidratio"
        - name: OTEL_TRACES_SAMPLER_ARG
          value: "0.1"
```

</details>
