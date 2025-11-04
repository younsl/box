[Timeslice](https://docs.newrelic.com/kr/docs/data-apis/understand-data/metric-data/query-apm-metric-timeslice-data-nrql/) 데이터를 사용해서 초당 GC 발생 카운트[ops/s] 계산하기

```bash
SELECT rate(sum(newrelic.timeslice.value), 5 minute) AS 'GC ops/s'
FROM Metric
WHERE k8s.clusterName = <CLUSTER_NAME>
AND k8s.podName LIKE '<POD_NAME>%'
AND k8s.namespaceName NOT IN('<EXCLUDED_NAMESPACE>')
AND (metricTimesliceName = 'GC/G1 Young Generation' OR metricTimesliceName = 'GC/G1 Concurrent GC')
FACET k8s.podName
```
