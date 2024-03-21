# 베이스 이미지 설정
FROM alpine:3.18

# 메타데이터 라벨 추가
LABEL maintainer="Younsung Lee <cysl@kakao.com>"
LABEL version="1.0"
LABEL description="Docker image for running Hugo blog server"
LABEL runtime="docker-desktop"

# 환경변수 설정
# ARG는 ENV와 다르게 build 타이밍에만 사용되며, 이미지에 포함되지 않습니다.
ARG HOME_DIR=/Users/younsung.lee

RUN apk update && \
    apk add --no-cache \
        hugo \
        git

# 작업 디렉토리 설정
WORKDIR /app

# 컨테이너 내부에 마운트할 로컬 디렉토리 설정
VOLUME $HOME_DIR/github/younsl/younsl.github.io:/app

# 포트 열기 (Hugo 서버 기본 포트)
EXPOSE 1313

# 컨테이너 실행 시 실행할 명령어
CMD ["hugo", "server", "-t", "etch", "--bind", "0.0.0.0"]
