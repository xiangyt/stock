# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 项目信息
BINARY_NAME=stock
BINARY_UNIX=$(BINARY_NAME)_unix
DOCKER_IMAGE=stock:latest

# 构建目录
BUILD_DIR=build

.PHONY: all build clean test coverage deps lint run-server run-cli run-worker docker-build docker-run init-db migrate docs help

# 默认目标
all: clean deps test build

# 构建所有二进制文件
build:
	@echo "Building binaries..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/server ./cmd/server
	$(GOBUILD) -o $(BUILD_DIR)/cli ./cmd/cli
	$(GOBUILD) -o $(BUILD_DIR)/worker ./cmd/worker

# 构建Linux版本
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) ./cmd/server

# 清理构建文件
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 测试覆盖率
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 运行服务器
run-server:
	@echo "Starting server..."
	$(GOCMD) run ./cmd/server

# 运行CLI工具
run-cli:
	@echo "Running CLI..."
	$(GOCMD) run ./cmd/cli

# 运行Worker
run-worker:
	@echo "Starting worker..."
	$(GOCMD) run ./cmd/worker

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Docker运行
docker-run:
	@echo "Running Docker container..."
	docker-compose up -d

# 初始化数据库
init-db:
	@echo "Initializing database..."
	$(GOCMD) run ./cmd/cli init-db

# 数据库迁移
migrate:
	@echo "Running database migration..."
	$(GOCMD) run ./cmd/cli migrate

# 生成API文档
docs:
	@echo "Generating API documentation..."
	swag init -g ./cmd/server/main.go -o ./docs

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/swaggo/swag/cmd/swag

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build        - Build all binaries"
	@echo "  build-linux  - Build for Linux"
	@echo "  clean        - Clean build files"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  run-server   - Run web server"
	@echo "  run-cli      - Run CLI tool"
	@echo "  run-worker   - Run background worker"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  init-db      - Initialize database"
	@echo "  migrate      - Run database migration"
	@echo "  docs         - Generate API documentation"
	@echo "  install-tools- Install development tools"
	@echo "  help         - Show this help message"