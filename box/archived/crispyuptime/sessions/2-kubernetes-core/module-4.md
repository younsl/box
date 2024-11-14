# 모듈 4 - 앱 업데이트

- **난이도**: 초급  
- **예상 소요 시간**: 10분

&nbsp;

이 시나리오의 목표는 `kubectl set image` 명령으로 배포된 애플리케이션을 업데이트하고, `rollout undo` 명령을 사용해 롤백하는 방법을 배우는 것입니다.

&nbsp;

## 1단계 - 네임스페이스 설정

모든 작업에서 반복적으로 `-n bootcamp`를 사용하지 않도록 기본 작업 네임스페이스를 `bootcamp`로 설정합니다. 먼저 `bootcamp`라는 네임스페이스가 생성되어 있는지 확인한 후, 아래 명령어로 기본 네임스페이스를 설정합니다:

```bash
kubectl config set-context --current --namespace=bootcamp
```

```console
$ kubectl config get-contexts multinode-lab
CURRENT   NAME            CLUSTER         AUTHINFO        NAMESPACE
*         multinode-lab   multinode-lab   multinode-lab   bootcamp
```

이제부터 명령어에서 `-n bootcamp`를 따로 지정하지 않아도 bootcamp 네임스페이스에서 작업이 이루어집니다.

&nbsp;

## 2단계 - 애플리케이션 버전 업데이트

먼저, 배포를 나열하려면 `get deployments` 명령을 실행합니다:

```bash
kubectl get deployments
```

실행 중인 파드를 나열하려면 get pods 명령을 실행합니다:

```bash
kubectl get pods
```

애플리케이션의 현재 이미지 버전을 확인하려면 `describe pods` 명령을 실행하고 Image 필드를 확인합니다:

```bash
kubectl describe pods
```

애플리케이션의 이미지를 버전 2로 업데이트하려면 set image 명령을 사용하고, 배포 이름과 새로운 이미지 버전을 지정합니다:

```bash
kubectl set image deployments/kubernetes-bootcamp kubernetes-bootcamp=gcr.io/k8s-minikube/kubernetes-bootcamp:v2
```

이 명령은 배포에 애플리케이션의 다른 이미지를 사용하도록 알리고 롤링 업데이트를 시작합니다.

```bash
deployment.apps/kubernetes-bootcamp image updated
```

`get pod` 명령을 사용해 새로운 파드의 상태를 확인하고, 기존 파드가 종료되는 모습을 확인합니다:

```bash
kubectl get pod -w
```

> [!NOTE]
> **Kubernetes 롤링 업데이트 개념**  
> Kubernetes는 애플리케이션을 중단 없이 업데이트할 수 있도록 롤링 업데이트 방식을 제공합니다. 롤링 업데이트는 기존 파드를 단계적으로 교체해, 애플리케이션을 지속적으로 제공하면서 업데이트를 수행할 수 있습니다. 이 방식은 애플리케이션의 가용성을 보장하는 중요한 Kubernetes 기능입니다.

&nbsp;

## 3단계 - 업데이트 확인

먼저 애플리케이션이 실행 중인지 확인합니다. 노출된 IP와 포트를 확인하려면 describe service 명령을 실행합니다:

```bash
kubectl describe services kubernetes-bootcamp
```

**Docker Desktop 사용자 참고**: Docker Desktop의 네트워크 제한으로 인해 기본적으로 호스트에서 파드에 직접 접근할 수 없습니다. `minikube service kubernetes-bootcamp` 명령을 실행하여 SSH 터널을 생성하고 브라우저에서 서비스를 확인할 수 있습니다. 터널을 종료하려면 Ctrl+C를 누르고, 이후 `curl $(minikube ip):$NODE_PORT` 명령을 실행해 튜토리얼을 계속 진행합니다.

&nbsp;

Node에 할당된 포트 값을 가지는 `NODE_PORT`라는 환경 변수를 생성합니다:

```bash
export NODE_PORT=$(kubectl get services/kubernetes-bootcamp -o go-template='{{(index .spec.ports 0).nodePort}}')
echo NODE_PORT=$NODE_PORT
```

다음으로 노출된 IP와 포트를 대상으로 `curl` 명령을 실행합니다:

```bash
curl $(minikube ip):$NODE_PORT
```

`curl` 명령을 실행할 때마다 다른 파드에 요청을 전송하게 됩니다. 모든 파드가 최신 버전(v2)을 실행 중임을 확인할 수 있습니다.

또한, `rollout status` 명령을 실행해 업데이트 상태를 확인할 수 있습니다:

```bash
kubectl rollout status deployments/kubernetes-bootcamp
```

애플리케이션의 현재 이미지 버전을 확인하려면 `describe pod` 명령을 실행합니다:

```bash
kubectl describe pods
```

출력의 Image 필드에서 최신 이미지 버전(v2)이 실행 중인지 확인합니다.

> [!NOTE]
> **Kubernetes의 배포 상태 관리**  
> Kubernetes는 배포(Deployment)의 상태를 추적하고, 새로운 버전이 배포되면 rollout을 통해 배포가 완료될 때까지 상태를 모니터링합니다. 이를 통해 관리자는 배포 프로세스를 실시간으로 확인하고, 배포 상태에 이상이 있을 경우 빠르게 조치할 수 있습니다.

&nbsp;

## 4단계 - 업데이트 롤백

이제 또 다른 업데이트를 수행하고 `v10` 태그가 붙은 컨테이너 이미지를 배포해 보겠습니다:

```bash
kubectl set image deployments/kubernetes-bootcamp kubernetes-bootcamp=gcr.io/k8s-minikube/kubernetes-bootcamp:v10
```

`get deployment` 명령을 사용해 배포 상태를 확인합니다:

```bash
kubectl get deployments
```

출력에 원하는 수의 파드가 나열되지 않은 것을 확인할 수 있습니다. get pods 명령을 사용해 모든 파드를 나열해 봅니다:

```bash
kubectl get pods
```

일부 파드가 `ImagePullBackOff` 상태인 것을 확인할 수 있습니다.

```console
$ kubectl get pods -w
NAME                                   READY   STATUS             RESTARTS   AGE
kubernetes-bootcamp-85f6fbf54c-ccx2h   0/1     ErrImagePull       0          21s
kubernetes-bootcamp-85f6fbf54c-kk2l4   0/1     ImagePullBackOff   0          21s
kubernetes-bootcamp-85f6fbf54c-m2tnd   0/1     ImagePullBackOff   0          20s
kubernetes-bootcamp-b5bd44ff6-2hnlq    1/1     Running            0          3m2s
kubernetes-bootcamp-b5bd44ff6-c5w9z    1/1     Terminating        0          2m55s
kubernetes-bootcamp-b5bd44ff6-ksdwb    1/1     Running            0          2m55s
kubernetes-bootcamp-b5bd44ff6-q2wqn    1/1     Running            0          3m2s
kubernetes-bootcamp-b5bd44ff6-w2kkk    1/1     Running            0          3m2s
```

문제에 대한 더 자세한 정보를 얻기 위해 describe pods 명령을 실행합니다:

```bash
kubectl describe pods
```

영향을 받은 파드의 출력의 Events 섹션에서 v10 이미지 버전이 레포지토리에 존재하지 않는다는 사실을 확인할 수 있습니다.

```console
$ kubectl describe pods

...

Events:
  Type     Reason          Age                From               Message
  ----     ------          ----               ----               -------
  Normal   Scheduled       38s                default-scheduler  Successfully assigned bootcamp/kubernetes-bootcamp-85f6fbf54c-m2tnd to multinode-lab
  Normal   SandboxChanged  33s                kubelet            Pod sandbox changed, it will be killed and re-created.
  Normal   BackOff         31s (x3 over 33s)  kubelet            Back-off pulling image "gcr.io/k8s-minikube/kubernetes-bootcamp:v10"
  Warning  Failed          31s (x3 over 33s)  kubelet            Error: ImagePullBackOff
  Normal   Pulling         16s (x2 over 38s)  kubelet            Pulling image "gcr.io/k8s-minikube/kubernetes-bootcamp:v10"
  Warning  Failed          13s (x2 over 33s)  kubelet            Failed to pull image "gcr.io/k8s-minikube/kubernetes-bootcamp:v10": Error response from daemon: manifest for gcr.io/k8s-minikube/kubernetes-bootcamp:v10 not found: manifest unknown: Failed to fetch "v10"
  Warning  Failed          13s (x2 over 33s)  kubelet            Error: ErrImagePull
```

이제 마지막으로 정상적으로 작동하던 버전으로 배포를 롤백하려면 rollout undo 명령을 사용합니다:

```bash
kubectl rollout undo deployments/kubernetes-bootcamp
```

rollout undo 명령은 배포를 이전에 정상적으로 작동하던 상태(v2 이미지)로 되돌립니다.

```bash
deployment.apps/kubernetes-bootcamp rolled back
```

Kubernetes는 업데이트를 버전으로 관리하며, 이전에 정상 작동했던 모든 상태로 되돌릴 수 있습니다.

get pods 명령을 다시 실행해 파드 목록을 확인합니다:

```bash
kubectl get pods
```

5개의 `bootcamp` 파드가 `bootcamp` 네임스페이스에서 실행 중입니다. 이 파드들에 배포된 이미지를 확인하려면 describe pods 명령을 실행합니다.

```bash
kubectl describe pods
```

배포는 다시 안정적인 애플리케이션 버전(v2)을 사용하고 있음을 확인할 수 있습니다. 롤백이 성공적으로 완료되었습니다.

> [!NOTE]
> **Kubernetes 롤백 기능**  
> Kubernetes는 배포를 관리할 때 이전의 상태를 기억하고, 문제가 발생할 경우 쉽게 롤백할 수 있는 기능을 제공합니다. 이를 통해 애플리케이션의 가용성과 신뢰성을 보장하며, 빠른 복구가 가능합니다.
