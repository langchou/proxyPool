# 变量定义
BINARY_NAME=proxypool
BUILD_DIR=build
DATA_DIR=$(BUILD_DIR)/data
GO_FILES=$(shell find . -name "*.go")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date +%FT%T%z)

# 默认目标
.DEFAULT_GOAL := build

# 创建必要的目录
.PHONY: init
init:
	mkdir -p $(BUILD_DIR)
	mkdir -p $(DATA_DIR)
	mkdir -p $(BUILD_DIR)/logs

# 清理构建目录
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# 编译
.PHONY: build
build: clean init
	# 编译程序
	go build -ldflags "-X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	# 复制配置文件
	cp config/config.toml $(DATA_DIR)/

# 运行
.PHONY: run
run: build
	cd $(BUILD_DIR) && ./$(BINARY_NAME)

# 测试
.PHONY: test
test:
	go test -v ./...

# 帮助
.PHONY: help
help:
	@echo "Make targets:"
	@echo "  build    - Build the application"
	@echo "  clean    - Remove build directory"
	@echo "  run      - Build and run the application"
	@echo "  test     - Run tests"
	@echo "  help     - Show this help" 