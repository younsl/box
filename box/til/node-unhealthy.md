Node NotReady:

- [Monitor Node Health](https://kubernetes.io/docs/tasks/debug/debug-cluster/monitor-node-health)
- [Draino and Node Problem Detector](https://gist.github.com/StevenACoffman/120bdbe8506e45bccc79bc73187c00bc)
  - **Warning**: Node Problem Detector only detects and reports node issues. You need a remedy system like [descheduler](https://github.com/kubernetes-sigs/descheduler) or [Draino](https://github.com/planetlabs/draino) to actually take corrective actions (pod eviction, node cordoning, etc.).
  - **Not Recommended**: [Draino](https://github.com/planetlabs/draino) is no longer maintained (last update December 2020) and is strongly not recommended for use.
- [Descheduler issue](https://github.com/kubernetes-sigs/descheduler/issues/131)
