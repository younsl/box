# Performance Optimization Lessons Learned

## Case Study: PerformanceOptimizer Anti-Pattern

### Background
CoCD의 PerformanceOptimizer는 GHES 서버 보호를 위해 도입되었으나, 실제로는 사용자 경험을 해치는 역효과를 일으켰습니다.

**백프레셔(Backpressure)**: 시스템이 과부하를 방지하기 위해 요청 처리 속도를 의도적으로 늦추는 기법입니다. 서버 응답이 느려지면 클라이언트가 요청 간격을 늘려 서버 부하를 줄이는 것이 목적입니다.

### Problem: Backpressure Over-Engineering

#### Original Implementation
```go
// 문제가 된 로직
if avgResponseTime > 2*time.Second {
    delay = baseDelay * 3  // 지연시간 3배 증가
} else if avgResponseTime > 1*time.Second {
    delay = baseDelay * 2  // 지연시간 2배 증가
}
```

#### What Went Wrong
1. **낮은 임계값**: 2초를 "느림"으로 판단하여 지연 증가
2. **과도한 배수**: 3배, 2배 증가로 급격한 성능 저하
3. **악순환 구조**: 느려질수록 더 느리게 만드는 피드백 루프

### Real-World Impact

#### Before Optimization (Expected)
- Worker 0: 500ms 지연
- Worker 1: 700ms 지연
- 예측 가능한 일정한 성능

#### After Optimization (Actual)
- GHES 응답시간 3.2초 발생
- Worker 0: 500ms × 3 = 1.5초 지연
- Worker 1: 700ms × 3 = 2.1초 지연
- 전체 스캔 시간 3배 이상 증가

### Anti-Patterns Identified

#### 1. Premature Optimization
```go
// 실제로는 불필요했던 복잡한 로직
type PerformanceOptimizer struct {
    responseTimeHistory []time.Duration
    avgResponseTime     time.Duration
    // ... 복잡한 상태 관리
}
```

**교훈**: 간단한 고정 지연이 더 효과적이었음

#### 2. Aggressive Backpressure
```go
// 너무 공격적인 백프레셔
SlowServerDelayMultiplier = 3  // 3배 증가는 과도함
```

**교훈**: 1.2~1.5배 정도의 온건한 증가가 적절

#### 3. Low Threshold Values
```go
// 너무 낮은 임계값
SlowResponseThreshold = 2 * time.Second  // GHES에서는 너무 낮음
```

**교훈**: 서버 특성을 고려한 현실적 임계값 설정 필요

#### 4. Unpredictable Behavior
```go
// 예측 불가능한 동적 지연
adaptiveDelay := optimizer.GetOptimalDelay(baseDelay)
// 상황에 따라 500ms~1.5초로 변동
```

**교훈**: 디버깅과 운영을 위해 예측 가능한 동작이 중요

### Recommended Approaches

#### Simple Fixed Delays
```go
// 단순하고 예측 가능한 지연
delay := BaseWorkerDelay + time.Duration(workerID)*WorkerDelayIncrement
// Worker 0: 500ms, Worker 1: 700ms (항상 일정)
```

#### Conservative Backpressure (if needed)
```go
// 온건한 백프레셔 (필요시에만)
if avgResponseTime > 10*time.Second {  // 높은 임계값
    delay = baseDelay * 1.2            // 낮은 배수
}
```

#### Circuit Breaker Pattern
```go
// 지속적인 장애 시 회로차단기
if consecutiveFailures > 5 {
    return ErrServiceUnavailable
}
```

### Guidelines

#### Before Optimization

1. **측정 먼저**: 실제 성능 문제 측정
2. **근본 원인 분석**: 최적화 대신 근본 원인 해결 고려
3. **단순함 우선**: 복잡한 최적화보다 단순한 해결책 선호
4. **점진적 적용**: 작은 변화부터 시작하여 점진적 개선

#### Red Flags

- **Magic Numbers**: 하드코딩된 임계값과 배수
- **Complex State**: 여러 상태를 관리하는 복잡한 구조
- **Feedback Loops**: 성능 저하가 더 큰 성능 저하를 유발
- **Unpredictable Timing**: 상황에 따라 크게 달라지는 타이밍

#### Testing Strategy

1. **A/B 테스트**: 최적화 전후 성능 비교
2. **부하 테스트**: 다양한 부하 상황에서 테스트
3. **장애 시나리오**: 서버 느림/장애 상황에서 동작 확인
4. **사용자 경험**: 실제 사용자 관점에서 성능 평가

### Conclusion

성능 최적화는 양날의 검입니다. 잘못 설계된 최적화는 성능을 개선하기보다 해칠 수 있습니다. 

**핵심 교훈**: 
- 단순함이 복잡함을 이긴다
- 측정 없는 최적화는 위험하다  
- 예측 가능성이 성능보다 중요할 수 있다
- 사용자 경험을 항상 고려하라

---

*이 문서는 CoCD PerformanceOptimizer 제거 과정에서 얻은 실제 교훈을 기록한 것입니다. (2025-07-21)*