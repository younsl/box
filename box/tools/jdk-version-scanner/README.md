# JDK Version Scanner

## Overview

This tool scans pods within specified Kubernetes namespaces (defaults to `default`) to identify installed Java Development Kit (JDK) versions. It filters out pods managed by DaemonSets.

## Usage

**(Optional)** Modify the `targetNamespaces` constant in `main.go` to target specific namespaces (comma-separated).

Run the script:

```bash
go run main.go
```

## Output Example

The script outputs a table listing the pods and their detected Java versions, followed by a summary.

```bash
Checking pods in namespace: default
Scanning pod 1 of 97: Checking Java version for pod example-pod-67c49d5ff8-sxm2z
...
INDEX  NAMESPACE  POD                           JAVA_VERSION
1      default    example-pod-67c49d5ff8-sxm2z  21.0.2
2      default    example-pod-6974df45d-sv4cq   17.0.4.1
...    ...        ...                           ...
97     default    example-pod-7b64fc9986-ncmh6  17.0.9
Total pods scanned: 343
Pods using JDK: 97
Time taken: 2m 34s
```