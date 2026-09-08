KITEX ?= kitex
IDL := idl/v0_1/shot_preview.thrift
LLM_IDL := idl/v0_1/llm_gateway.thrift
ASSET_IDL := idl/v0_1/asset_service.thrift

SHELL := /bin/bash
NODE_BIN_PATH := $(HOME)/.nvm/versions/node/$(shell [ -d $(HOME)/.nvm/versions/node ] && ls -1 $(HOME)/.nvm/versions/node 2>/dev/null | tail -n 1)/bin

.DEFAULT_GOAL := dev

.PHONY: generate test run server web dev start stop check-env install-web build clean

generate:
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(IDL)
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(LLM_IDL)
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(ASSET_IDL)

test:
	go test ./...

stop:
	@echo "==> 检查并停止占用端口 8888, 8889, 3000 的既有服务进程..."
	@-lsof -ti :8888 | xargs kill -9 2>/dev/null || true
	@-lsof -ti :8889 | xargs kill -9 2>/dev/null || true
	@-lsof -ti :3000 | xargs kill -9 2>/dev/null || true
	@echo "==> 端口已成功清理与释放。"

check-env:
	@echo "==> 检查开发环境完整性..."
	@which go >/dev/null 2>&1 || (echo "[错误] 未检测到 Go 环境，请先安装 Go 1.22+" && exit 1)
	@export PATH="$(NODE_BIN_PATH):$(PATH)"; \
	which node >/dev/null 2>&1 || (echo "[错误] 未检测到 Node.js，请先安装 Node.js (推荐 v18+)" && exit 1); \
	which npm >/dev/null 2>&1 || (echo "[错误] 未检测到 npm，请先安装 npm" && exit 1)
	@if [ ! -d "web/node_modules" ] || [ ! -d "web/node_modules/three" ] || [ "web/package.json" -nt "web/node_modules" ]; then \
		echo "==> 检测到前端依赖缺失或 package.json 已更新，正在自动执行 npm install..."; \
		export PATH="$(NODE_BIN_PATH):$(PATH)"; \
		cd web && npm install && touch node_modules; \
	else \
		echo "==> 前端依赖完整 (web/node_modules 已就绪)"; \
	fi
	@echo "==> 环境健康检查通过！"

install-web:
	@export PATH="$(NODE_BIN_PATH):$(PATH)"; \
	cd web && npm install

server: stop
	@echo "==> 正在启动后端服务 (Kitex RPC: 8889, HTTP Gateway: 8888)..."
	go run ./cmd/server

web:
	@export PATH="$(NODE_BIN_PATH):$(PATH)"; \
	cd web && npm run dev

# 一键启动全栈服务：自动停止旧进程 -> 检查并自动补齐依赖 -> 并发启动前后端 -> Ctrl+C 联动退出
dev: stop check-env
	@echo ""
	@echo "=========================================================================="
	@echo "🚀 正在一键启动 Blender Shot Preview 全栈服务..."
	@echo "   - 后端 API: http://127.0.0.1:8888 (RPC: 127.0.0.1:8889)"
	@echo "   - 前端 Web: http://localhost:3000"
	@echo "   - 退出说明: 按 Ctrl+C 将自动停止前后端所有服务进程"
	@echo "=========================================================================="
	@echo ""
	@export PATH="$(NODE_BIN_PATH):$(PATH)"; \
	trap 'echo -e "\n==> 正在优雅关闭前后端进程..."; kill 0; exit 0' INT TERM; \
	go run ./cmd/server & \
	(cd web && npm run dev) & \
	wait

start: dev

run: dev

build:
	@echo "==> 编译检查后端..."
	go build -o /dev/null ./cmd/server
	@echo "==> 构建检查前端..."
	@export PATH="$(NODE_BIN_PATH):$(PATH)"; \
	cd web && npm run build
	@echo "==> 全栈构建校验通过！"
