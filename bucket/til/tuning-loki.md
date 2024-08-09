# loki tuning

## 문제점

기본값으로 설치된 `loki-distributed`에서 `ingester` 파드가 간헐적으로 10vCPU에 12GB 메모리를 쓰면서 터진다.

<img width="1000px" alt="image" src="https://github.com/user-attachments/assets/c2712bf7-2edd-4730-910f-b2b5768f6dc1">

## ingester 설정

### storage의 최대 조회기간

`max_look_back_period`는 Loki 인제스터가 검색 쿼리 시 검색할 수 있는 최대 기간을 정의합니다. 이 값이 설정되면, 사용자 쿼리가 이 기간을 넘는 데이터에 대해서는 검색할 수 없습니다.

예를 들어, `max_look_back_period`가 720시간(30일)으로 설정되면, 쿼리가 30일보다 오래된 데이터에 대해 검색할 수 없게 됩니다. 이는 데이터의 양이 너무 많아져서 인제스터가 과도한 리소스를 소비하지 않도록 방지합니다.

```yaml
loki:
  config: |
    chunk_store_config:
      max_look_back_period: 720h
```

### Replication Factor

```yaml
# charts/loki-distributed/values.yaml
loki:
  config: |
    ingester:
      lifecycler:
        ring:
          kvstore:
            store: memberlist
          replication_factor: 3
```

Loki ingester의 RF가 3인 경우, 3개 파드 모두 동일한 데이터를 가지고 있기 때문에 rollingUpdate 발생시에도 데이터의 내구성과 가용성도 잘 유지됩니다. PDB도 추가하도록 합니다.

```yaml
# charts/loki-distributed/values.yaml
ingester:
  maxUnavailable: 1
```

### WAL(Write Ahead Logging)

WAL을 사용하면 Loki의 로그 데이터가 디스크에 기록되기 전에 메모리에서의 변경사항을 기록합니다. 결과적으로 `ingester` 파드 일부가 유실되더라도 데이터는 보존됩니다.

![black_background_image](https://github.com/user-attachments/assets/8a8e150b-7b94-4e37-b912-ce06dbfa17d8)

이는 Loki 2.2 릴리스부터 사용할 수 있는 새로운 기능으로, Write를 허용하기 전에 Loki가 들어오는 모든 데이터를 디스크<sup>`/var/loki/wal`</sup>에 기록하여, 로그가 손실되지 않도록 하는 데 도움이 됩니다.

```yaml
# charts/loki-distributed/values.yaml
loki:
  config: |
    ingester:
      wal:
        dir: /var/loki/wal
```

기본적으로 `loki-distributed` 차트에서 `wal` 볼륨은 비활성화되어 있습니다.

```yaml
# charts/loki-distributed/values.yaml
ingester:
  persistence:
    # -- Enable creating PVCs which is required when using boltdb-shipper
    enabled: true
    # -- Use emptyDir with ramdisk for storage. **Please note that all data in ingester will be lost on pod restart**
    inMemory: false
    # -- List of the ingester PVCs
    # @notationType -- list
    claims:
      - name: data
        size: 30Gi
        #   -- Storage class to be used.
        #   If defined, storageClassName: <storageClass>.
        #   If set to "-", storageClassName: "", which disables dynamic provisioning.
        #   If empty or set to null, no storageClassName spec is
        #   set, choosing the default provisioner (gp2 on AWS, standard on GKE, AWS, and OpenStack).
        storageClass: null
      # - name: wal
      #   size: 150Gi
```

## 참고자료

- [The essential config settings you should use so you won’t drop logs in Loki](https://grafana.com/blog/2021/02/16/the-essential-config-settings-you-should-use-so-you-wont-drop-logs-in-loki/)
- [TSDB Birds Eye View](https://lokidex.com/posts/tsdb/)
