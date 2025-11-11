## 개요

New Relic 데이터비용 절감 포인트를 찾기 위한 튜닝 가이드

## 튜닝 가이드

### Tracing

Span 데이터 크기(GB) 계산:

```sql
SELECT bytecountestimate()/10e8 as 'Span Estimate'
FROM Span
SINCE 1 day ago
```

```bash
1323
```

```bash
Tracing = Span + ErrorTrace + SqlTrace
```

```sql
SELECT bytecountestimate()/10e8 as 'Tracing Estimate'
FROM Span, ErrorTrace, SqlTrace SINCE 1 day ago
```

```bash
1324
```

```bash
(1323 / 1324) * 100 = 99.92%
```

Span 데이터가 전체 트레이싱의 99.92%를 차지하고 있음

Top 10 Span 데이터 차지 비율

```bash
SELECT bytecountestimate()/10e8
FROM Span
SINCE 1 day ago FACET entity.name LIMIT 10
```

OpenTelemetry가 수집되는 데이터가 많은 이유:

- Newrelic은 Span, Transaction 등을 샘플링합니다.
  - Adaptive Sampling을 사용하며, Default로 에이전트 당 1분에 최대 2000개의 Span을 수집
  - 1분에 200개 이상의 Span이 수집되는 APM 에이전트 수에 비례해 Ingest가 늘어남
- OpenTelemetry의 경우도 Sampling은 하지만 방식이 약간 다릅니다.
  - Sampler가 전부/Paranet Span 여부 기반/확률적으로 수집되는 Span의 양을 조절함
  - 에이전트 개수(앱 인스턴스의 개수)와 상관없이 생성되는 Span을 전부 Or 일정 비율로 수집하므로 App에 들어오는 리퀘스트의 개수에 비례해서 수집되는 span의 수도 상승함