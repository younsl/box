# secrets-store-csi-driver

## 한계점

Secrets Store CSI Driver에서 Kubernetes Secret은 Secret을 마운트한 파드가 시작된 이후에만 동기화(생성)됩니다. 반대로 파드가 삭제되면 Kubernetes Secret도 삭제됩니다. 자세한 사항은 [Sync as Kubernetes Secret](https://secrets-store-csi-driver.sigs.k8s.io/topics/sync-as-kubernetes-secret.html) 문서를 참고합니다.

이 부분이 싫다면 [External Secrets Operator](https://external-secrets.io/) 혹은 [HashiCorp Vault](https://www.vaultproject.io/)를 고려해보세요.