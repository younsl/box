# 빗썸 Cloud Engineer 면접 Q&A

## 1. AWS & 클라우드 인프라

### Q: AWS에서 블록체인 서비스를 운영한 경험을 설명해주세요.
**A:** 
- EKS에서 Ethereum 노드를 StatefulSet으로 운영, PVC로 블록체인 데이터 영구 저장
- NLB를 통한 RPC 엔드포인트 노출, AWS WAF로 Rate Limiting 적용
- CloudWatch + Prometheus로 노드 동기화 상태, 피어 연결 수, 블록 높이 모니터링
- Auto Scaling Group으로 노드 가용성 보장, 멀티 AZ 배포로 고가용성 확보

### Q: 24/7 서비스 운영 경험과 장애 대응 프로세스를 설명해주세요.
**A:**
- **모니터링 체계**: Datadog APM으로 실시간 트랜잭션 추적, PagerDuty로 온콜 관리
- **장애 대응**: RTO 30분, RPO 5분 목표로 운영, Runbook 기반 표준화된 대응
- **실제 사례**: RDS 커넥션 풀 고갈 시 즉시 파라미터 그룹 수정 및 재시작 없이 적용
- **사후 관리**: Post-mortem 문서화, 재발 방지를 위한 자동화 스크립트 개발

### Q: 보안 관점에서 클라우드 인프라를 어떻게 관리하셨나요?
**A:**
- **네트워크 보안**: Private Subnet 구성, Security Group 최소 권한 원칙, NACLs 설정
- **접근 제어**: AWS SSM Session Manager로 Bastion 없이 안전한 접근, MFA 필수화
- **암호화**: KMS로 EBS/RDS 암호화, Secrets Manager로 API 키 관리
- **컴플라이언스**: AWS Config Rules로 규정 준수 자동 검증, GuardDuty로 위협 탐지

## 2. Kubernetes & 컨테이너

### Q: Kubernetes에서 블록체인 노드를 운영하는 베스트 프랙티스는?
**A:**
```yaml
# StatefulSet 예시
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ethereum-node
spec:
  serviceName: ethereum
  replicas: 3
  volumeClaimTemplates:
  - metadata:
      name: blockchain-data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: gp3
      resources:
        requests:
          storage: 2Ti
```
- StatefulSet으로 안정적인 네트워크 ID와 스토리지 보장
- PodDisruptionBudget으로 업그레이드 시 가용성 유지
- Node Affinity로 고성능 인스턴스에 배치
- Init Container로 체인 데이터 스냅샷 복원

### Q: Kubernetes 운영 중 겪은 어려움과 해결 방법은?
**A:**
- **문제**: 노드 OOM으로 인한 Pod Eviction
- **해결**: Resource Limits/Requests 적절히 설정, VPA로 자동 조정
- **문제**: etcd 성능 저하
- **해결**: etcd 전용 노드 분리, SSD 스토리지 사용, 정기적 defragmentation

## 3. CI/CD & 자동화

### Q: GitOps 기반 CI/CD 파이프라인을 구축한 경험을 설명해주세요.
**A:**
```yaml
# GitHub Actions 예시
name: Deploy to EKS
on:
  push:
    branches: [main]
jobs:
  deploy:
    steps:
    - uses: actions/checkout@v3
    - name: Build and Push to ECR
      run: |
        aws ecr get-login-password | docker login --username AWS --password-stdin $ECR_REGISTRY
        docker build -t $ECR_REPOSITORY:$IMAGE_TAG .
        docker push $ECR_REPOSITORY:$IMAGE_TAG
    - name: Update Kubernetes Manifest
      run: |
        kubectl set image deployment/app app=$ECR_REPOSITORY:$IMAGE_TAG
```
- ArgoCD로 Git 저장소와 Kubernetes 클러스터 동기화
- Helm Chart로 환경별 값 관리, Sealed Secrets로 민감 정보 암호화
- Blue-Green, Canary 배포 전략 구현

### Q: IaC 도구를 활용한 인프라 관리 경험은?
**A:**
- **Terraform**: AWS 리소스 프로비저닝, 모듈화로 재사용성 향상
- **상태 관리**: S3 Backend + DynamoDB Lock으로 협업 환경 구성
- **실제 구현**: VPC, EKS, RDS, ElastiCache 등 전체 인프라 코드화
```hcl
module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  cluster_name    = "blockchain-cluster"
  cluster_version = "1.28"
  
  node_groups = {
    blockchain = {
      desired_capacity = 3
      instance_types   = ["r5.2xlarge"]
      disk_size       = 1000
    }
  }
}
```

## 4. 블록체인 특화

### Q: 블록체인 노드 운영 시 주의할 점은?
**A:**
- **데이터 관리**: 체인 데이터 증가 대비 스토리지 자동 확장 설정
- **네트워크**: 피어 연결 수 최적화, 대역폭 모니터링
- **동기화**: 초기 동기화 시간 단축을 위한 스냅샷 활용
- **보안**: Private Key 관리를 위한 HSM 또는 KMS 활용

### Q: 스마트 컨트랙트 배포 파이프라인을 어떻게 구성하시겠습니까?
**A:**
1. **개발**: Hardhat/Truffle로 로컬 테스트
2. **테스트**: Testnet 자동 배포, Integration 테스트
3. **감사**: Slither/Mythril로 보안 취약점 자동 스캔
4. **배포**: Multi-sig 지갑으로 Mainnet 배포
5. **검증**: Etherscan API로 컨트랙트 자동 검증

## 5. 모니터링 & 로깅

### Q: 블록체인 서비스의 모니터링 전략은?
**A:**
```yaml
# Prometheus 메트릭 예시
- job_name: 'ethereum'
  metrics_path: '/metrics'
  static_configs:
  - targets: ['geth:6060']
  metric_relabel_configs:
  - source_labels: [__name__]
    regex: 'eth_.*'
    action: keep
```
- **인프라 메트릭**: CPU, Memory, Disk I/O, Network
- **블록체인 메트릭**: 블록 높이, 피어 수, 트랜잭션 풀 크기
- **비즈니스 메트릭**: TPS, API 응답 시간, 에러율
- **알림**: 블록 지연, 노드 동기화 실패, 메모리 누수 감지

### Q: 로그 수집 및 분석 파이프라인은?
**A:**
- **수집**: Fluent Bit으로 컨테이너 로그 수집
- **저장**: OpenSearch에 인덱싱, S3에 장기 보관
- **분석**: Kibana 대시보드로 실시간 분석
- **알림**: ElastAlert로 이상 패턴 감지 시 Slack 알림

## 6. 협업 & 소프트 스킬

### Q: 개발팀과 협업하여 문제를 해결한 경험은?
**A:**
- **상황**: API 응답 지연으로 사용자 불만 증가
- **분석**: APM으로 DB 쿼리 병목 발견
- **협업**: 개발팀과 함께 쿼리 최적화, 인덱스 추가
- **결과**: 응답 시간 80% 감소, Redis 캐싱 레이어 추가로 추가 개선

### Q: 새로운 기술 도입 경험과 프로세스는?
**A:**
- **평가**: PoC 진행, 기존 스택과의 호환성 검토
- **도입**: 단계적 롤아웃, 파일럿 프로젝트 선정
- **교육**: 팀 내 지식 공유 세션, 문서화
- **예시**: Service Mesh(Istio) 도입으로 마이크로서비스 간 통신 보안 강화

## 7. 기술 역량 체크

### Q: 현재 사용 가능한 기술 스택은?
**A:**
- **클라우드**: AWS (EKS, RDS, Lambda, CloudFormation)
- **컨테이너**: Docker, Kubernetes, Helm
- **CI/CD**: GitHub Actions, ArgoCD, Jenkins
- **IaC**: Terraform, Ansible
- **모니터링**: Prometheus, Grafana, Datadog
- **언어**: Python, Go, Bash
- **협업**: JIRA, Confluence, Slack

### Q: 블록체인 관련 기술 이해도는?
**A:**
- **합의 알고리즘**: PoW, PoS, PBFT 이해
- **네트워크**: P2P 통신, 노드 발견 메커니즘
- **암호화**: 공개키/개인키, 해시 함수, 머클 트리
- **스마트 컨트랙트**: Solidity 기본, Gas 최적화 이해

## 8. 상황별 문제 해결

### Q: 긴급 장애 상황 시 대응 프로세스는?
**A:**
1. **즉시 대응**: 서비스 영향도 파악, 임시 조치 (롤백/스케일아웃)
2. **원인 분석**: 로그 분석, 메트릭 확인, 재현 테스트
3. **영구 조치**: 근본 원인 해결, 코드/설정 수정
4. **사후 관리**: RCA 작성, 재발 방지 대책 수립

### Q: 비용 최적화를 위한 노력은?
**A:**
- **Reserved/Spot Instance** 활용으로 EC2 비용 60% 절감
- **S3 Lifecycle Policy**로 오래된 데이터 Glacier 이동
- **Cost Explorer**로 월별 비용 분석, 태깅으로 부서별 비용 할당
- **Karpenter**로 Kubernetes 노드 자동 스케일링 최적화

## 9. 빗썸 특화 질문

### Q: 암호화폐 거래소의 인프라 요구사항은?
**A:**
- **고가용성**: 99.99% 이상 가동률, 무중단 배포
- **보안**: 콜드/핫 월렛 분리, 다중 서명, HSM 활용
- **성능**: 초당 수만 건 거래 처리, 마이크로초 단위 레이턴시
- **규정 준수**: KYC/AML 데이터 관리, 감사 로그 보관

### Q: 빗썸에 기여할 수 있는 부분은?
**A:**
- 클라우드 네이티브 아키텍처로 확장성 개선
- GitOps 도입으로 배포 자동화 및 안정성 향상
- 모니터링 고도화로 장애 예방 및 빠른 대응
- 비용 최적화로 운영 효율성 증대

## 10. 마무리 질문

### Q: 빗썸에 궁금한 점은?
**준비할 역질문:**
1. 현재 사용 중인 기술 스택과 향후 도입 계획은?
2. 팀 구성과 협업 방식은 어떻게 되나요?
3. 온콜 근무 체계와 장애 대응 프로세스는?
4. 블록체인 노드 운영 규모와 종류는?
5. 성장 기회와 교육 지원은 어떻게 되나요?

---

## 면접 준비 팁

1. **기술 깊이**: 각 기술의 원리와 트러블슈팅 경험 준비
2. **블록체인 이해**: 기본 개념과 거래소 특성 파악
3. **보안 마인드**: 금융 서비스의 보안 중요성 강조
4. **협업 능력**: 개발팀과의 소통 경험 준비
5. **열정 표현**: 블록체인과 기술 성장에 대한 관심 어필

화이팅! 🚀