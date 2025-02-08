# dind cgroup issue

## 개요

cgroup v2 환경에서 dind 컨테이너 안에서 실행되는 컨테이너가 너무 많은 메모리를 사용하는 문제가 있었다. Pod limit을 초과하는 문제였다.

![cgroup v2](./assets/1.png)

dind 컨테이너에서 실행되는 컨테이너가 Pod limit 16GB를 인식하지 못하고, 노드 전체 메모리 32GB를 자신의 메모리로 인식하여 메모리를 모두 사용하는 문제임.

dind에서 실행된 컨테이너가 노드 전체 메모리를 사용중인 모습:

![dind](./assets/2.png)

관련 이슈:

- [Control Group v2](https://docs.kernel.org/admin-guide/cgroup-v2.html)
- [wrong container_memory_working_set_bytes for DinD container on cgroupv2 only nodes #119853](https://github.com/kubernetes/kubernetes/issues/119853#issuecomment-1938042934) : dind 컨테이너 메트릭 표기 문제
- [DinD cgroupv2 problem inside K8s #45378](https://github.com/moby/moby/issues/45378) : dind 컨테이너 리소스 문제
- [Allow us to set Docker's Memory Usage #540](https://github.com/game-ci/unity-builder/issues/540#issuecomment-2248602115) : dind 컨테이너 메모리 제한 문제