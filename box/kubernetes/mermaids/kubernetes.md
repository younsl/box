# kubernetes

## Summary

System diagram written in [Mermaid](https://mermaid.js.org/). All diagrams are related to kubernetes.

## Mermaid diagrams

```mermaid
---
title: kubelet and ALB health check
---
flowchart LR
    subgraph "Kubernetes Cluster"
        subgraph "Worker Node"
            kubelet("kubelet")
            pod("`**Pod**
            (Your App)`")
        end
        
        kubelet -- "Port 8081<br/>/actuator/..." --> pod
    end

    alb("`**ALB**
        Internet-facing`")

    alb -- "Port 8080<br/>/actuator/..." --> pod
    client("Client") --> alb

    style alb fill:darkorange, color:white, stroke:#333, stroke-width:2px
    linkStyle 1 stroke:red,stroke-width:2px
```