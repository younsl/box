# Job Tracker

[![Go Version](https://img.shields.io/badge/go-1.25-000000?style=flat-square&logo=go&logoColor=white)](go.mod)
[![GitHub release](https://img.shields.io/github/v/release/younsl/box?style=flat-square&color=black&logo=github&logoColor=white&label=release)](https://github.com/younsl/box/releases?q=job-tracker)
[![License](https://img.shields.io/github/license/younsl/box?style=flat-square&color=black&logo=github&logoColor=white)](/LICENSE)

구직 활동을 체계적으로 관리하고 추적하는 웹 애플리케이션

## 주요 기능

### 핵심 기능
- 지원 현황 관리 - 회사별 지원 내역 추가/수정/삭제
- 상태 추적 - 지원완료/서류심사/인터뷰/오퍼/탈락/철회 단계별 관리
- 파일 첨부 - 이력서, 자소서 등 관련 문서 업로드 및 관리
- 필터링 및 검색 - 상태별, 회사별 빠른 검색
- 통계 대시보드 - 지원 현황 한눈에 보기

### 보안 및 동기화
- GPG 암호화 - 모든 데이터를 GPG로 암호화하여 저장
- SQLite 데이터베이스 - 경량화된 로컬 데이터 저장소
- Git 동기화 - 암호화된 데이터를 Git으로 백업 및 버전 관리
- 로컬 전용 - 외부 서버 없이 완전한 로컬 환경에서 동작

### UI/UX
- 모던 미니멀 디자인 - 깔끔한 모노톤 인터페이스
- 반응형 웹 - 모바일, 태블릿, 데스크톱 모두 지원
- 실시간 업데이트 - 즉각적인 UI 반영

## 기술 스택

- **Backend**: Go 1.25
- **Frontend**: HTML5, TailwindCSS, Alpine.js
- **Database**: SQLite3
- **Security**: GPG (GnuPG)
- **Build**: Make, Air (hot-reload)
- **Container**: Docker

## 설치 및 실행

### 사전 요구사항

- Go 1.25 이상
- GPG (GnuPG) 2.0 이상
- Git
- Make

### 빠른 시작

#### 1. GPG 설정

**⚠️ 중요 경고: GPG 개인키를 분실하면 암호화된 데이터를 절대 복구할 수 없습니다!**
- 반드시 GPG 개인키를 안전한 곳에 백업하세요
- 패스프레이즈를 잊어버리면 데이터 복구가 불가능합니다
- 키 백업 없이 시스템을 재설치하면 모든 데이터를 잃게 됩니다

```bash
# GPG 키 확인
gpg --list-secret-keys --keyid-format LONG

# GPG 키가 없다면 새로 생성
gpg --full-generate-key
# - RSA and RSA 선택
# - 키 크기: 4096
# - 유효기간: 0 (무제한) 또는 원하는 기간
# - 실명과 이메일 입력

# ⚠️ 필수: GPG 개인키 백업
gpg --export-secret-keys -a "your-email@example.com" > private-key-backup.asc
# 이 파일을 USB, 외장하드 등 안전한 곳에 보관하세요!
```

#### 2. 환경 변수 설정

```bash
# .env 파일 생성 또는 쉘에 export
export GPG_RECIPIENT="your-email@example.com"  # GPG 키에 등록된 이메일
export PORT=1314                                # 웹 서버 포트 (기본값: 1314)
export DB_PATH="./data.db"                      # SQLite DB 경로 (기본값: ./data.db)
```

#### 3. 빌드 및 실행

```bash
# 저장소 클론
git clone https://github.com/yourusername/job-tracker.git
cd job-tracker

# 의존성 설치 및 초기 설정
make setup

# 개발 모드 실행 (자동 리로드)
make dev

# 프로덕션 빌드 및 실행
make build
make run

# Docker로 실행
make docker-build
make docker-run
```

#### 4. 웹 브라우저에서 접속

```
http://localhost:1314
```

## 사용 방법

### 새 지원 추가

1. "New Application" 버튼 클릭
2. 회사명, 포지션, 지원 플랫폼 입력
3. 지원 URL 추가 (선택사항)
4. 관련 파일 첨부 (이력서, 자소서 등)
5. "Save" 클릭

### 상태 업데이트

1. 지원 목록에서 회사 클릭
2. 상태 드롭다운에서 현재 진행 상태 선택
   - `applied` - 지원 완료
   - `screening` - 서류 심사 중
   - `interview` - 인터뷰 진행
   - `offer` - 오퍼 받음
   - `rejected` - 탈락
   - `withdrawn` - 지원 철회
3. 메모 섹션에 상세 내용 기록
4. "Update" 클릭

### 파일 관리

1. 지원 상세 페이지에서 "Attach Files" 클릭
2. 파일 선택 (최대 10MB)
3. 파일 설명 추가
4. 업로드된 파일은 암호화되어 저장

### Git 동기화

```bash
# 자동 동기화 (웹 UI)
웹 인터페이스에서 "Sync with Git" 버튼 클릭

# 수동 동기화
make sync

# 데이터 백업
make backup

# 데이터 복원
make restore
```

## 데이터 구조

### SQLite 데이터베이스 스키마

```sql
-- 지원 정보 테이블
CREATE TABLE applications (
    id TEXT PRIMARY KEY,
    company TEXT NOT NULL,
    position TEXT NOT NULL,
    status TEXT NOT NULL,
    final_result TEXT,
    applied_date TEXT NOT NULL,
    platform TEXT,
    url TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 첨부파일 테이블
CREATE TABLE attachments (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    description TEXT,
    file_data BLOB,           -- 파일 데이터를 BLOB으로 저장
    file_size INTEGER,
    mime_type TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);
```

### JSON 데이터 구조 (GPG 암호화)

```json
{
  "applications": [
    {
      "id": "1234567890",
      "company": "회사명",
      "position": "포지션",
      "status": "applied",
      "final_result": "",
      "applied_date": "2024-01-20",
      "platform": "LinkedIn",
      "url": "https://...",
      "notes": "1차 인터뷰 통과",
      "files": [
        {
          "filename": "resume_2024.pdf",
          "description": "최신 이력서",
          "size": 245632
        },
        {
          "filename": "cover_letter.pdf",
          "description": "자기소개서",
          "size": 128456
        }
      ],
      "last_modified": "2024-01-20T15:30:00Z"
    }
  ],
  "last_updated": "2024-01-20T15:30:00Z"
}
```

### 파일 저장 구조

```
job-tracker/
├── data.db                 # SQLite 데이터베이스 (첨부파일 BLOB 포함)
├── data.db.gpg            # 암호화된 데이터베이스 (첨부파일 포함 전체 암호화)
└── backups/              # 백업 디렉토리
    ├── backup_20240120.db.gpg
    └── ...
```

**중요**: 첨부파일은 별도 디렉토리가 아닌 SQLite 데이터베이스의 `attachments` 테이블에 BLOB으로 저장됩니다. 
전체 데이터베이스가 GPG로 암호화되므로 첨부파일도 함께 암호화됩니다.

## 보안

### GPG 암호화

**⚠️ 치명적 주의사항: GPG 개인키 분실 = 모든 데이터 영구 손실**

이 애플리케이션은 GPG 암호화를 사용하여 데이터를 보호합니다. 
**GPG 개인키나 패스프레이즈를 분실하면 어떠한 방법으로도 데이터를 복구할 수 없습니다.**

- 모든 민감한 데이터는 GPG로 암호화
- 데이터베이스 백업 시 자동 암호화 (첨부파일 포함)
- SQLite DB 전체를 하나의 파일로 암호화
- Git 저장소에는 암호화된 DB 파일만 커밋
- **복호화는 오직 원본 GPG 개인키로만 가능**

### 보안 모범 사례

1. **GPG 키 관리 (매우 중요)**
   - **개인키 백업은 필수** - 분실 시 데이터 영구 손실
   - 강력한 패스프레이즈 사용 (최소 20자 이상 권장)
   - 개인키 백업본을 여러 안전한 장소에 보관
   - 패스프레이즈를 안전한 곳에 별도 기록
   - 정기적인 키 백업 및 복구 테스트 수행
   
   ```bash
   # 개인키 백업
   gpg --export-secret-keys -a "your-email@example.com" > gpg-secret-backup.asc
   
   # 개인키 복구
   gpg --import gpg-secret-backup.asc
   ```

2. **Git 저장소**
   - Private 저장소 사용 권장
   - `.gitignore`에 민감한 파일 추가
   - 암호화되지 않은 데이터 커밋 방지

3. **로컬 보안**
   - 디스크 암호화 활성화
   - 정기적인 백업
   - 안전한 네트워크에서만 사용

## API 엔드포인트

### Applications

- `GET /api/applications` - 전체 지원 목록 조회
- `GET /api/applications/:id` - 특정 지원 상세 조회
- `POST /api/applications` - 새 지원 추가
- `PUT /api/applications/:id` - 지원 정보 수정
- `DELETE /api/applications/:id` - 지원 삭제

### Attachments

- `GET /api/applications/:id/files` - 첨부파일 목록
- `POST /api/applications/:id/files` - 파일 업로드
- `GET /api/files/:fileId` - 파일 다운로드
- `DELETE /api/files/:fileId` - 파일 삭제

### System

- `POST /api/sync` - Git 동기화
- `POST /api/backup` - 데이터 백업
- `POST /api/restore` - 데이터 복원
- `GET /api/stats` - 통계 정보

## 개발

### 프로젝트 구조

```
job-tracker/
├── cmd/
│   └── job-tracker/
│       └── main.go        # 애플리케이션 진입점
├── pkg/
│   ├── app/              # 애플리케이션 로직
│   ├── config/           # 설정 관리
│   ├── crypto/           # GPG 암호화
│   ├── git/              # Git 동기화
│   ├── logging/          # 로깅
│   ├── models/           # 데이터 모델
│   └── storage/          # 저장소 인터페이스
├── web/
│   ├── static/           # 정적 파일 (CSS, JS)
│   └── templates/        # HTML 템플릿
├── Dockerfile            # 컨테이너 이미지
├── Makefile             # 빌드 및 실행 스크립트
└── go.mod               # Go 모듈 정의
```

### 개발 명령어

```bash
# 테스트 실행
make test

# 코드 포맷팅
make fmt

# 린트 검사
make lint

# 벤치마크
make bench

# 클린 빌드
make clean
make build

# 도커 이미지 빌드
make docker-build

# 도커 컴포즈 실행
make docker-compose-up
```

### 환경 변수

| 변수명 | 설명 | 기본값 |
|--------|------|--------|
| `PORT` | 웹 서버 포트 | 1314 |
| `DB_PATH` | SQLite DB 경로 | ./data.db |
| `GPG_RECIPIENT` | GPG 수신자 이메일 | (필수) |
| `LOG_LEVEL` | 로그 레벨 (debug/info/warn/error) | info |
| `BACKUP_DIR` | 백업 디렉토리 | ./backups |
| `MAX_FILE_SIZE` | 최대 파일 크기 (bytes) | 10485760 (10MB) |
| `AUTO_BACKUP` | 자동 백업 활성화 | true |
| `BACKUP_INTERVAL` | 백업 주기 (hours) | 24 |

## 문제 해결

### GPG 관련 문제

```bash
# GPG 에이전트 재시작
gpgconf --kill gpg-agent
gpg-agent --daemon

# 키 신뢰도 설정
gpg --edit-key your-email@example.com
> trust
> 5 (ultimate)
> quit

# 권한 문제 해결
chmod 700 ~/.gnupg
chmod 600 ~/.gnupg/*
```

### 데이터베이스 문제

```bash
# DB 무결성 검사
sqlite3 data.db "PRAGMA integrity_check;"

# DB 복구
sqlite3 data.db ".dump" | sqlite3 data_new.db
mv data_new.db data.db

# 마이그레이션 실행
make migrate
```

### 동기화 문제

```bash
# Git 상태 확인
git status

# 충돌 해결
git pull --rebase
# 충돌 파일 수정 후
git add .
git rebase --continue

# 강제 동기화 (주의!)
make force-sync
```

## 기여하기

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일 참조
