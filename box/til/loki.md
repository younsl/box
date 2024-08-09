
https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/#rate-limit-errors

```
limits_config:
  ingestion_rate_mb: 4
  ingestion_burst_size_mb: 6

  ingestion-rate_mb: 20
  ingestion_burst_size_mb: 40
# Per-user ingestion rate limit in sample size per second. Units in MB.
# CLI flag: -distributor.ingestion-rate-limit-mb
[ingestion_rate_mb: <float> | default = 4]

# Per-user allowed ingestion burst size (in sample size). Units in MB. The burst
# size refers to the per-distributor local rate limiter even in the case of the
# 'global' strategy, and should be set at least to the maximum logs size
# expected in a single push request.
# CLI flag: -distributor.ingestion-burst-size-mb
[ingestion_burst_size_mb: <float> | default = 6]
```
