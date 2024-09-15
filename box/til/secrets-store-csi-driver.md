# secrets-store-csi-driver

Secrets Store CSI Driver는 Kubernetes Secret이 생성되는 시점은 Secret을 마운트하는 파드를 시작한 후에만 동기화되는 치명적인 한계점이 있음.

이 부분이 싫다면 External Secrets Operator를 고려해보자.
