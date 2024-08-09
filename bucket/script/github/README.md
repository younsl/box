# github

Github Cloud와 Github Enterprise Server에서 사용하는 자동화 스크립트들

## 스크립트 현황

> [!NOTE]
> Platform type 값에서 **GHC**는 Github Cloud, **GHEC**는 Github Enterprise Cloud를, **GHES**는 Github Enterprise Server (self-hosted)를 의미합니다.

| Platform type | Script name | Language | Description |
|---------------|-------------|----------|-------------|
| `GHC` | commit-history-cleaner.sh | bash | Commit log(history) 전체 삭제 |
| `GHC` | remove-workflows-run.sh | bash | Workflow 실행기록 전체 삭제 |
| `GHES` | scan-schedule-workflow | python | `schedule`로 cron trigger 되는 모든 Workflow 목록과 실행 시간을 스캐닝 |
