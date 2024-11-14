# 실습 2: StatefulSet을 사용하여 데이터베이스 배포하기

## 배경지식

### StatefulSet

[StatefulSet](https://kubernetes.io/ko/docs/concepts/workloads/controllers/statefulset/)은 Kubernetes에서 상태가 있는 애플리케이션을 배포하고 관리하기 위해 사용되는 리소스입니다. 이는 데이터베이스와 같은 상태 저장 애플리케이션을 지원하며, 각 Pod가 고유한 정체성을 가지고, 안정적인 네트워크 ID와 영속적인 스토리지 볼륨을 갖도록 합니다.

&nbsp;

### StatefulSet의 주요 특징:

1. **고유한 정체성**: 각 Pod는 고유한 이름을 가지며, 이 이름은 StatefulSet의 이름을 기반으로 하여 생성됩니다. 예를 들어, `mysql-0`, `mysql-1`, `mysql-2`와 같은 형태입니다.
2. **영속적 스토리지**: 각 Pod는 고유한 Persistent Volume을 갖고 있어 Pod가 재시작되더라도 데이터가 보존됩니다.
3. **순서 보장**: Pod의 생성, 업데이트, 삭제가 순서대로 이루어집니다. 새로운 Pod가 생성될 때 기존 Pod가 준비가 되어야만 다음 Pod가 생성됩니다.
4. **Stable Network Identity**: 각 Pod는 DNS를 통해 안정적인 네트워크 주소를 가지며, 이는 Pod 간의 통신을 용이하게 합니다.

&nbsp;

### StatefulSet을 사용하는 이유:

- 상태 저장<sup>Stateful</sup> 애플리케이션을 손쉽게 관리할 수 있습니다.
- 데이터가 손실되지 않도록 안정적인 스토리지를 제공합니다.
- 복잡한 상태 저장 애플리케이션을 배포할 때 유용합니다.

&nbsp;

## 목표

minikube 클러스터에 StatefulSet을 사용하여 MySQL 데이터베이스를 배포해보겠습니다.

## 실습

노드 3개로 구성된 `minikube` 클러스터를 생성합니다.

```bash
minikube start \
    --driver='docker' \
    --cni='calico' \
    --kubernetes-version='stable' \
    --nodes=3
```

&nbsp;

컨트롤 플레인과 워커 노드의 상태를 확인합니다.

```bash
$ kubectl get node
NAME           STATUS   ROLES           AGE   VERSION
minikube       Ready    control-plane   80s   v1.31.0
minikube-m02   Ready    <none>          56s   v1.31.0
minikube-m03   Ready    <none>          36s   v1.31.0
```

&nbsp;

이 시나리오에서는 쿠버네티스 상에서 DB를 구성할 것입니다. 데이터베이스는 3개의 복제본을 가지며, 각 복제본은 고유한 Persistent Volume을 가져야 합니다.

&nbsp;

StatefulSet 구성 파일 작성하기: [mysql-statefulset.yaml 파일](./1-sts-mysql.yaml)을 작성합니다.

StatefulSet 배포하기:

```yaml
wget https://raw.githubusercontent.com/younsl/crispyuptime/main/sessions/3-handson/1-sts-mysql.yaml
kubectl apply -f 1-sts-mysql.yaml
```

&nbsp;

배포된 `mysql` 파드 상태를 확인합니다.

```bash
kubectl get pod -w
```

Pod들이 순차적으로 `mysql-0` → `mysql-1` → `mysql-2` 순서로 생성됩니다. 여기서 우리는 Statefulset으로 제어되는 파드들은 순서가 보장된다는 사실을 알 수 있습니다.

`kubectl get pod -w`의 출력 결과:

```bash
NAME      READY   STATUS              RESTARTS   AGE
mysql-0   1/1     Running             0          30s
mysql-1   0/1     ContainerCreating   0          14s
mysql-1   1/1     Running             0          18s
mysql-2   0/1     Pending             0          0s
mysql-2   0/1     Pending             0          0s
mysql-2   0/1     Pending             0          2s
mysql-2   0/1     ContainerCreating   0          2s
mysql-2   0/1     ContainerCreating   0          3s
mysql-2   1/1     Running             0          19s
```

&nbsp;

이제 StatefulSet의 순서 보장을 확인하기 위해 모든 Pod를 삭제하고 재시작하는 과정을 실습합니다.

모든 Pod 삭제:

```bash
kubectl delete pod -l app=mysql \
  && kubectl get pods -w
```

Pod 순서 확인: Pod들이 `mysql-0` → `mysql-1` → `mysql-2` 순서로 다시 생성됩니다.

```bash
$ kubectl get pod -w
NAME      READY   STATUS    RESTARTS   AGE
mysql-0   1/1     Running   0          4s
mysql-1   1/1     Running   0          3s
mysql-2   1/1     Running   0          2s
```

`rollout` 명령어로 재시작도 해봅니다.

```bash
kubectl rollout restart statefulset mysql \
  && kubectl get pod -w
```

이번에는 같이 가장 최근에 생성된 파드부터 `mysql-2` → `mysql-1` → `mysql-0` 역순으로 재시작하게 됩니다.

```bash
statefulset.apps/mysql restarted
NAME      READY   STATUS        RESTARTS   AGE
mysql-0   1/1     Running       0          2m3s
mysql-1   1/1     Running       0          107s
mysql-2   1/1     Terminating   0          89s
mysql-2   1/1     Terminating   0          89s
mysql-2   0/1     Completed     0          89s
mysql-2   0/1     Completed     0          90s
mysql-2   0/1     Completed     0          90s
mysql-2   0/1     Pending       0          0s
mysql-2   0/1     Pending       0          0s
mysql-2   0/1     ContainerCreating   0          0s
mysql-2   0/1     ContainerCreating   0          1s
mysql-2   1/1     Running             0          1s
mysql-1   1/1     Terminating         0          109s
mysql-1   1/1     Terminating         0          110s
mysql-1   0/1     Completed           0          110s
mysql-1   0/1     Completed           0          111s
mysql-1   0/1     Completed           0          111s
mysql-1   0/1     Pending             0          0s
mysql-1   0/1     Pending             0          0s
mysql-1   0/1     ContainerCreating   0          0s
mysql-1   0/1     ContainerCreating   0          0s
mysql-1   1/1     Running             0          1s
mysql-0   1/1     Terminating         0          2m8s
mysql-0   1/1     Terminating         0          2m9s
mysql-0   0/1     Completed           0          2m9s
mysql-0   0/1     Completed           0          2m10s
mysql-0   0/1     Completed           0          2m10s
mysql-0   0/1     Pending             0          0s
mysql-0   0/1     Pending             0          0s
mysql-0   0/1     ContainerCreating   0          0s
mysql-0   0/1     ContainerCreating   0          0s
mysql-0   1/1     Running             0          1s
```

&nbsp;

~~파드를 DB로 돌리면, 언제든 종료될 수 있다는 의미가 아닐까?~~ 미친짓은 아닙니다.

- [쿠버네티스 기반 데이터베이스 운영 도전기 #1 - 사례 조사편](https://sungchul-p.github.io/db-on-k8s-case-study)

&nbsp;

볼륨의 이해:

여기서 PROVISIONER가 `kubernetes.io/host-path`인 것을 확인할 수 있습니다. 이는 호스트(사용자 PC 또는 가상머신)의 파일 시스템을 사용함을 의미합니다.

```console
$ kubectl get sc
NAME                 PROVISIONER                RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   k8s.io/minikube-hostpath   Delete          Immediate           false                  24m
```

&nbsp;

StatefulSet이 정상적으로 배포되었는지 확인합니다.

```bash
kubectl get statefulsets
kubectl get sts
```

&nbsp;

Pod 및 Persistent Volume 확인하기:

```bash
kubectl get pods
kubectl get pvc
```

&nbsp;

각 Pod와 Persistent Volume Claim이 생성되었는지 확인합니다.

&nbsp;

```bash
kubectl exec -n default -it mysql-0 -- mysql -u root -p
Enter password: password
```

&nbsp;

MySQL에 성공적으로 접속하면, MySQL 프롬프트가 나타납니다. 데이터베이스 목록을 확인하려면 다음 명령어를 실행합니다.

```bash
mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
4 rows in set (0.01 sec)
```

```bash
CREATE DATABASE testdb;
SHOW DATABASES;
USE testdb;

CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255));

INSERT INTO users (name) VALUES ('Alice'), ('Bob');

SELECT * FROM users;
```

&nbsp;

정상적으로 `users` 테이블에 데이터가 들어갔다.

```bash
+----+-------+
| id | name  |
+----+-------+
|  1 | Alice |
|  2 | Bob   |
+----+-------+
2 rows in set (0.00 sec)
```

이제 MySQL Pod을 삭제하여 재시작한 후에도 데이터가 유지되는지를 확인해 보겠습니다. 첫 번째 `mysql` Pod을 삭제합니다.

```console
$ kubectl delete pod mysql-0
pod "mysql-0" deleted
```

&nbsp;

Kubernetes는 자동으로 mysql-0 Pod를 다시 생성합니다. 이 일련의 과정은 아래 명령어로 한눈에 확인 가능합니다. 

```bash
kubectl delete pod mysql-0 -n default \
  && kubectl get pod -w
```

Pod가 준비되면 다음 명령어로 데이터가 유지되는지 확인합니다.

```
kubectl exec -it mysql-0 -- mysql -u root -p
```

```bash
mysql> USE testdb;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> SELECT * FROM users;
+----+-------+
| id | name  |
+----+-------+
|  1 | Alice |
|  2 | Bob   |
+----+-------+
2 rows in set (0.00 sec)
```

파드가 삭제된 후에도 Alice와 Bob의 데이터가 여전히 존재합니다. 이는 Persistent Volume을 사용해 데이터가 영구 보존된 덕분입니다.

&nbsp;

## 결론

- Minikube는 기본적으로 standard StorageClass를 제공하여 PVC와 로컬 볼륨 간의 바인딩을 자동화합니다.
- StatefulSet을 사용하면 Pod가 삭제되거나 재시작되어도 데이터가 유지됩니다.

이 실습을 통해 StatefulSet과 StorageClass의 연계 동작을 이해하고, 상태 저장 애플리케이션 배포 방법을 배울 수 있습니다.

&nbsp;

### 스테이트풀셋의 사용 사례

StatefulSet은 주로 다음과 같은 사례에 사용됩니다.

- 데이터베이스(MySQL, PostgreSQL)
- 분산 스토리지 시스템(Cassandra, Elasticsearch)
- 메시지 브로커(Kafka, RabbitMQ)
- 애플리케이션 캐시 서버(Redis, Memcached)

위 사례들은 모두 상태 저장이 중요한 시스템이며, 이러한 애플리케이션은 데이터의 무결성, 노드 간 순서 보장, 세션 일관성 유지가 필수적이기 때문에 쿠버네티스 상에 배포하는 경우 StatefulSet이 가장 적합한 선택입니다.
