---
title: "kubecost session"
date: 2024-11-27T11:02:10+09:00
lastmod: 2024-11-27T11:02:15+09:00
slug: ""
description: "kubecost"
keywords: []
tags: ["devops", "iac", "terraform", "terragrunt"]
---

## 컨테이너에 대한 성능과 비용 관리 방안

- 컨테이너에 대한 성능과 비용 관리 방안 (2024/11/27 11:10 ~ 11:40)

IBM KubeCost

---

구소망 차장, Client Engineering

- 컨테이너에 대한 성능과 비용 관리 방안을 주제로 정함

Agenda

- Kubernetes Challenge
- Kubecost Concept
- Kubecost Architecture
- kubecost Use Cases
- Why Kubecost?

---

- 84%가 쿠버네티스 도입을 검토, 사용중
- 기업의 72%는 컨테이너 비용 최적화 수행을 최소한으로만 하고 있어, 개선할 수 있는 충분한 기회가 있음

---

Kubernetes Cost Optimization Goal

팀, 애플리케이션, 프로젝트, 네트워크 등 모든 Kubernetes와 관련된 Cloud 사용 비용과 성능 관리가 증가함.

---

Kubecost Concept

- K8s 환경의 급속한 성장
- 팀 애플리케이션, 프로젝트 등 별로 K8s와 Cloud 비용을 할당하고 확인
- 블랙박스인 네트워크 비용의 가시화
- 더 간단한 자동화로 비용 절감
- K8s와 Cloud 지출을 보다 효과적으로 관리

---

UKubecost Use case

- Cost Allocation : 적절한 자원 할당
- Optimization Insight : 실행 가능한 통찰력
- Alerts & Governance : 사용자 정의 알림

---

Savings 탭에서 여러가지 비용 절감할 수 있는 포인트들을 요약해서 정리. 이를 적용하면 됨.

- PV right-sizing
- 사용하지 않는 PV 리소스 현황을 요약 (Orphan Resource)

---

Kubecost Use Cases > Alerts & Governance : Budgets

- 예산에 대한 지출을 모니터링
- 예산 초과 방지 알림
- 보고서 생성 가능 (예): 팀별 예산 관리, 프로젝트 기반 비용 통제

---

Why Kubecosst: 다양한 산업의 고객

- Coinbase
- GitLab
- Splunk
- Nvidia

---

[SKT TANGO 사례](https://www.sktenterprise.com/bizInsight/blogDetail/dev/3909)

- 대규모 Kubernetes 클러스터 운영: SKT는 통신망 관제 시스템은 TANGO를 구축하며 K8s 클러스터 운영하게 됨. 클라우드 급증으로 효율적인 관리가 필요했음
- 비용 최적화의 필요성: K8s 환경에서 발생하는 비용을 정확히 팡가하고, 최적화하기 위해 Kubecost를 도입했었음. 초기 단계부터 성능 및 비용 최적화를 목표로 했음
