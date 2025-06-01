.PHONY: build test clean docker-build docker-push deploy help

# 변수
BINARY_NAME=eip-rotation-handler
DOCKER_REPO?=eip-rotation-handler
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go 관련 변수
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# AWS 관련 변수 (환경변수에서 가져옴)
AWS_ACCOUNT_ID?=YOUR_ACCOUNT_ID
AWS_REGION?=us-west-2
IMAGE_URI=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(DOCKER_REPO):$(VERSION)

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

## help: 사용 가능한 명령어 표시
help:
	@echo "사용 가능한 명령어:"
	@echo "  build         - 바이너리 빌드"
	@echo "  test          - 테스트 실행"
	@echo "  clean         - 빌드 파일 정리"
	@echo "  docker-build  - Docker 이미지 빌드"
	@echo "  docker-push   - Docker 이미지 푸시"
	@echo "  deploy        - Kubernetes에 배포"
	@echo "  deps          - 의존성 정리"
	@echo "  version       - 버전 정보 표시"

## build: Go 바이너리 빌드
build: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

## test: 테스트 실행
test:
	$(GOTEST) -v ./...

## clean: 빌드 파일 정리
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

## deps: Go 모듈 의존성 정리
deps:
	$(GOMOD) tidy
	$(GOMOD) download

## docker-build: Docker 이미지 빌드
docker-build:
	docker build -t $(DOCKER_REPO):$(VERSION) .
	docker tag $(DOCKER_REPO):$(VERSION) $(IMAGE_URI)

## docker-push: Docker 이미지 ECR에 푸시
docker-push: docker-build
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
	docker push $(IMAGE_URI)

## deploy: Kubernetes에 배포
deploy:
	@echo "Deploying to Kubernetes..."
	chmod +x scripts/deploy.sh
	AWS_ACCOUNT_ID=$(AWS_ACCOUNT_ID) AWS_REGION=$(AWS_REGION) IMAGE_TAG=$(VERSION) ./scripts/deploy.sh

## dev: 개발용 로컬 실행
dev:
	LOG_LEVEL=debug ROTATION_INTERVAL_MINUTES=1 $(GOCMD) run ./cmd/$(BINARY_NAME)

## fmt: 코드 포맷팅
fmt:
	$(GOCMD) fmt ./...

## vet: 코드 정적 분석
vet:
	$(GOCMD) vet ./...

## lint: golangci-lint 실행 (설치 필요)
lint:
	golangci-lint run

## all: 전체 빌드 및 테스트
all: deps fmt vet test build

## version: 버전 정보 표시
version:
	@echo $(VERSION) 