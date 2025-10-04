# ArgoCD Application Controller

- [#6125](https://github.com/argoproj/argo-cd/issues/6125#issuecomment-3360729444)
  - ArgoCD의 application-controller는 기본적으로 Cluster-level sharding만 지원하며, 동일 클러스터의 Application-level sharding은 지원하지 않아, application-controller 파드가 여러개인 경우 균등한 부하분산이 이뤄지지 않음
