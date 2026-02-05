.PHONY: build clean test install run-server run-k8s run-all help

# 构建变量
BINARY_NAME=inspection-tool
VERSION?=1.0.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_TIME}"

help: ## 显示帮助信息
	@echo "可用的make目标:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@echo "编译 $(BINARY_NAME)..."
	go build ${LDFLAGS} -o $(BINARY_NAME) cmd/main.go
	@echo "✓ 编译完成: $(BINARY_NAME)"

build-linux: ## 编译Linux版本
	@echo "编译 Linux 版本..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(BINARY_NAME)-linux-amd64 cmd/main.go
	@echo "✓ 编译完成: $(BINARY_NAME)-linux-amd64"

build-all: ## 编译所有平台版本
	@echo "编译所有平台版本..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(BINARY_NAME)-linux-amd64 cmd/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o $(BINARY_NAME)-linux-arm64 cmd/main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(BINARY_NAME)-darwin-amd64 cmd/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o $(BINARY_NAME)-darwin-arm64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o $(BINARY_NAME)-windows-amd64.exe cmd/main.go
	@echo "✓ 所有版本编译完成"

clean: ## 清理编译文件
	@echo "清理编译文件..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf reports/
	@echo "✓ 清理完成"

test: ## 运行测试
	go test -v -race -coverprofile=coverage.out ./...

install: build ## 安装到系统
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✓ 安装完成"

deps: ## 下载依赖
	@echo "下载依赖..."
	go mod download
	go mod tidy
	@echo "✓ 依赖下载完成"

run-server: build ## 运行服务器巡检示例
	./$(BINARY_NAME) server --host 127.0.0.1 --user root --password changeme

run-k8s: build ## 运行K8s巡检示例
	./$(BINARY_NAME) k8s --kubeconfig ~/.kube/config

run-all: build ## 运行综合巡检示例
	./$(BINARY_NAME) all --kubeconfig ~/.kube/config --ssh-user root --ssh-password changeme

lint: ## 运行代码检查
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint" && exit 1)
	golangci-lint run ./...

fmt: ## 格式化代码
	go fmt ./...
	gofmt -s -w .

version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
