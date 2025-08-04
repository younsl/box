## 증상

CoreDNS v1.11.3-eksbuild.2 버전에서 발생

```yaml
# coredns spec
spec:
  containers:
    resources:
      requests:
        memory: 170Mi
      limits:
        cpu: 100m
        memory: 70Mi
```

CoreDNS 에러 로그:

```bash
2025-02-10T01:24:12.124250291Z coredns-8dbc687fd-ccr29 [ERROR] plugin/errors: 2 g.comail.com. A: read udp xxx.xxx.xxx.150:52166->xx.xx.0.2:53: i/o timeout
2025-02-10T01:24:20.582015168Z coredns-8dbc687fd-ccr29 [ERROR] plugin/errors: 2 g.comail.com. A: read udp xxx.xxx.xxx.150:48624->xx.xx.0.2:53: i/o timeout
2025-02-10T01:24:22.685874989Z coredns-8dbc687fd-ccr29 [ERROR] plugin/errors: 2 g.comail.com. AAAA: read udp xxx.xxx.xxx.150:53252->xx.xx.0.2:53: i/o timeout
2025-02-12T04:17:37.714371550Z coredns-8dbc687fd-ccr29 [ERROR] plugin/errors: 2 www.googleapis.com. A: read udp xxx.xxx.xxx.150:56386->xx.xx.0.2:53: i/o timeout
2025-02-12T04:17:37.714423281Z coredns-8dbc687fd-ccr29 [ERROR] plugin/errors: 2 www.googleapis.com. AAAA: read udp xxx.xxx.xxx.150:52035->xx.xx.0.2:53: i/o timeout
```

Amazon 제공 DNS 서버의 IP 주소는 `x.x.x.2`로 끝납니다. 예를 들어 VPC의 CIDR 범위가 `10.10.0.0/16`이라면, 이 범위 내의 모든 서브넷에서 사용하는 Amazon 제공 DNS(Amazon Provided DNS) 서버의 주소는 `10.10.0.2`이 됩니다.

간헐적인 `i/o timeout` 에러 발생

## 원인

CoreDNS 파드의 모든 CPU Request가 300%를 넘으며, sum(rate(process_cpu_seconds_total{instance=~"$instance"}[5m])) 메트릭의 결과값이 7.5 ~ 8.5초 정도를 상시 유지했었음

## 참고자료

[forward causing high cpu usage #4544](https://github.com/coredns/coredns/issues/4544): CPU Request가 300%를 넘으면 문제가 있을 수 있음