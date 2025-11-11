---
title: "rules"
date: 2025-01-16T12:20:20+09:00
lastmod: 2025-01-16T12:20:20+09:00
slug: ""
description: "Rulebook for Kubernetes best practice"
keywords: []
tags: ["aws", "kubernetes"]
---

## 개요

몇 년간 쿠버네티스 클러스터를 운영하면서 모범사례를 모아둔 모음입니다.

&nbsp;

## 모범사례

1. Pod의 CPU Limit은 설정하지 않는게 좋습니다. CPU Throttle 문제를 빈번하게 일으킵니다.
2. 외부 노출이 필요 없는 Service 리소스는 ClusterIP 타입으로 설정하여 외부 노출을 방지합니다.
3. 헬스체크 설정시 Instance Type 대신 IP Type으로 헬스체크를 하는 것이 좋습니다. Instance Type인 경우, 100개의 워커노드 EC2가 있는 경우 고작 1개의 파드를 위해 100배의 헬스체크가 발생합니다.
4. 쿠버네티스 클러스터에 Prometheus를 구성할 때 [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)을 사용하는 것이 좋습니다. 대부분의 빌트인 대시보드와 프로메테우스, 그라파나를 포함하고 있기 때문에 빠르게 구성할 수 있습니다.
5. 파드의 권한 획득은 워커노드 롤에 권한을 추가하는 대신 [IRSA(IAM Role for Service Account)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) 방식을 사용해 특정 파드만 해당 권한을 접근할 수 있게 구성하도록 하자.
