# coredns

## autoscaling

EKS 에드온으로 관리되는 `coredns`에 오토스케일링이 적용하면 [CPA](https://github.com/kubernetes-sigs/cluster-proportional-autoscaler)<sup>Cluster Proportional Autoscaler</sup>에 의해 `coredns` 파드 개수가 조절됩니다.

CoreDNS의 자동 크기 조정은 Amazon EKS에서 관리하는 클러스터 Control Plane의 새로운 구성 요소인 Cluster Proportional Autoscaler에 의해 수행됩니다.

결과적으로 `kubectl get hpa -n kube-system`으로 조회해도 CoreDNS 관련 HPA 리소스가 조회되지 않는게 맞습니다.

```bash
$ kubectl get hpa -A
NAMESPACE   NAME                 REFERENCE                       TARGETS                                     MINPODS   MAXPODS   REPLICAS   AGE
argocd      argocd-repo-server   Deployment/argocd-repo-server   memory: <unknown>/60%, cpu: <unknown>/60%   1         10        4          3h14m
argocd      argocd-server        Deployment/argocd-server        memory: <unknown>/80%, cpu: <unknown>/80%   1         5         5          3h14m
# ... There is no HPA resource for coredns ...
```

&nbsp;

EKS Module을 사용하여 EKS 클러스터를 관리하는 경우, 테라폼 코드에서 `configuration_values` 값을 통해 Autoscaling을 설정할 수 있습니다.

```json
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name                   = local.name
  cluster_version                = local.cluster_version
  cluster_endpoint_public_access = true

  cluster_addons = {
    coredns = {
      addon_version = "v1.11.1-eksbuild.9"
      configuration_values = jsonencode({
        autoScaling = {
          enabled     = true
          minReplicas = 2
          maxReplicas = 10
        }
      })
    }
  }
```

&nbsp;

EKS 콘솔에서 coredns 에드온의 Advanced configuration는 다음과 같이 반영됩니다.

```json
{
  "autoScaling": {
    "enabled": true,
    "maxReplicas": 10,
    "minReplicas": 2
  }
}
```

## 참고자료

[Autoscaling CoreDNS](https://docs.aws.amazon.com/eks/latest/userguide/coredns-autoscaling.html#coredns-autoscaling-prereqs)
