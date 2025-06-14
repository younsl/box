# Karpenter

## Summary

System diagram written in [Mermaid](https://mermaid.js.org/). All diagrams are related to Karpenter.

## Mermaid diagrams

### Figure 1

```mermaid
---
title: "Figure 1. Karpenter Node Provisioning"
---
flowchart LR
  subgraph "Kubernetes Cluster"
    direction LR
    subgraph "Node Provisioning Config"
      np["nodepool"]
      ec2nc["ec2nodeclass"]
    end
    
    subgraph "Karpenter Controller"
      kc["`**Pod**
      Karpenter Controller`"]
    end
    
    subgraph "Pending Pods"
      p1["`**Pod**
      (Pending)`"]
      p2["`**Pod**
      (Pending)`"]
    end

    nc["nodeclaim"]
    cp["`**Pod**
    kube-apiserver`"]
  end
  
  subgraph "AWS"
    wn["EC2 Instance<br/>(Worker Node)"]
  end

  kc --Watch pending pods--> p1 & p2
  kc --> np --> ec2nc
  kc --"Create nodeclaim resource"--> nc --> wn
  wn --"Join cluster"--> cp

  style np fill:darkorange,color:white,stroke:#333,stroke-width:2px
  style ec2nc fill:darkorange,color:white,stroke:#333,stroke-width:2px
```

&nbsp;

### Figure 2

```mermaid
---
title: "Figure 2. Scrape metrics from Karpenter"
---
flowchart LR
  subgraph "Kubernetes Cluster"
    direction LR
    subgraph "kube-system namespace"
      sm["serviceMonitor"]
      svc["`**Service**
      ClusterIP`"]
      kc["`**Pod**
      karpenter`"]
    end

    subgraph "monitoring namespace"
      prop["`**Pod**
      Prometheus Operator`"]
      prom["`**Pod**
      Prometheus`"]
    end

    prop --> sm --> svc e1@--/metrics--> kc

    prop --Update config--> prom

    prom e2@--Scrape metrics--> svc
  end

  style sm fill:darkorange,color:white,stroke:#333,stroke-width:2px
  e1@{ animate: true }
  e2@{ animate: true }
```

&nbsp;

### Figure 3

```mermaid
---
title: "Figure 3. Helm chart structure of Karpenter"
---
flowchart LR
  admin("👨🏻‍💼 Cluster Admin")
  subgraph "Cluster"
    subgraph "kube-system"
      subgraph "karpenter chart"
        kc["`karpenter
          Main chart`"]
        knpc["`karpenter-nodepool chart
              Subchart`"]
      end

      kcp["`**Pod**
            karpenter controller`"]
      np["`**Custom Resource**
          nodepool`"]
      ec2nc["`**Custom Resource**
              ec2nodeclass`"]
    end
  end

  admin --helm install--> kc & knpc
  kc --> kcp
  knpc --> np & ec2nc

  style kc fill:darkblue
  style knpc fill:darkblue
```