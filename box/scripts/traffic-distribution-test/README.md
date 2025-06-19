# Traffic Distribution Test

Real-time TCP connection monitoring tool for testing Kubernetes Service trafficDistribution Local AZ communication.

## Usage

```bash
sh conmon.sh
```

## Features

- Tests Service trafficDistribution Local AZ communication functionality
- Monitors TCP connections to ports 80/8080 to verify traffic locality
- Generates background traffic to the target service
- Displays connection count per IP address to verify Local AZ preference
- Runs 20 monitoring cycles with 2-second intervals

## Sample Output

```
=== Traffic Distribution Test ===
Service: web-api.production.svc.cluster.local
Testing Local AZ communication...
Press Ctrl+C to stop

CHECK    [COUNT]  IP_ADDRESS
----------------------------------------
1/20     [12]     172.20.143.85
2/20     [10]     172.20.143.85
3/20     [10]     172.20.143.85
4/20     [10]     172.20.143.85
5/20     [12]     172.20.143.85
...
20/20    [9]      172.20.143.85

Monitoring completed.
```

## Configuration

Edit the SERVICE_NAME variable in `conmon.sh` to monitor your target service:

```bash
SERVICE_NAME="your-service.namespace.svc.cluster.local"
```
