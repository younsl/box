# k8s-registry-io-stat

<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/kubernetes/kubernetes-plain.svg" width="40" height="40"/>

This script tests the availability of the Kubernetes registry by sending a series of HTTP requests and calculating the success rate.

The script was created in response to the issue documented at [GitHub Issue #7670](https://github.com/kubernetes/k8s.io/issues/7670), which discusses intermittent 503 errors when accessing the registry at `registry.k8s.io`.

## Usage

Set the `REQUEST_COUNT` environment variable to the number of requests you want to send.

```bash
REQUEST_COUNT=10
```

Run the script:

```bash
sh test-registry.sh
```
