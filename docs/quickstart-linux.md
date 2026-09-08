# Blender Shot Preview — Linux 快速启动指南

本文档帮助你在 Linux 系统上快速完成环境准备并启动 **Blender Shot Preview** 全栈服务（后端 Kitex RPC + HTTP Gateway，前端 React/Vite）。

---

## 一、前置条件

### 1. Go 1.22+

```bash
# 检查是否已安装
go version

# 若未安装，以 Go 1.22.4 为例（根据实际最新版本替换）
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz

# 写入 PATH（追加到 ~/.bashrc 或 ~/.zshrc）
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Node.js 18+ 与 npm

推荐使用 [nvm](https://github.com/nvm-sh/nvm) 管理 Node.js 版本：

```bash
# 安装 nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc

# 安装并启用 Node.js 18 LTS
nvm install 18
nvm use 18

# 验证
node -v   # 应输出 v18.x.x
npm -v
```

### 3. Git

```bash
# Debian/Ubuntu
sudo apt-get install -y git

# RHEL/CentOS/Fedora
sudo dnf install -y git

# 验证
git --version
```

### 4. make

```bash
# Debian/Ubuntu
sudo apt-get install -y make

# RHEL/CentOS/Fedora
sudo dnf install -y make
```

---

## 二、获取代码

```bash
git clone https://github.com/S-zhi/blender-shot-preview.git
cd blender-shot-preview
```

---

## 三、配置环境变量

后端启动**必须**设置 `LLM_GATEWAY_MASTER_KEY`，它是一个 32 字节的加密主密钥，以 64 位十六进制字符串传入。

### 生成密钥

```bash
# 使用系统随机源生成 32 字节密钥，并编码为十六进制
openssl rand -hex 32
# 输出示例：a3f2c1e4b5d6789012345678abcdef01fedcba9876543210a1b2c3d4e5f60718
```

### 设置方式（二选一）

**方式 A：当前会话临时生效（推荐快速测试）**

```bash
export LLM_GATEWAY_MASTER_KEY="<上面生成的64位十六进制字符串>"
```

**方式 B：写入项目根目录 `.env` 文件（推荐持久化开发）**

```bash
# 项目本身不自动加载 .env，需配合 direnv 或在启动前手动 source
echo 'export LLM_GATEWAY_MASTER_KEY="<64位十六进制字符串>"' > .env
source .env
```

> **注意**：请勿将 `.env` 或含真实密钥的文件提交到版本库。项目 `.gitignore` 已包含 `.env`，请确认后再操作。

> **重要**：密钥长度必须恰好为 64 位十六进制字符（对应 32 字节）。长度不符会导致服务启动失败。

---

## 四、安装依赖

```bash
# 下载 Go 模块依赖
go mod download

# 安装前端依赖
cd web && npm install && cd ..
```

如果网络访问 Go 模块代理较慢，可设置国内代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

---

## 五、一键启动（开发模式）

```bash
make dev
```

该命令会自动完成以下步骤：
1. 停止可能占用端口 8888、8889、3000 的旧进程
2. 检查 Go、Node.js、npm 环境完整性，缺少前端依赖时自动执行 `npm install`
3. 并发启动后端服务和前端开发服务器
4. 按 **Ctrl+C** 即可联动停止所有进程

---

## 六、服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端 Web | http://localhost:3000 | React 开发服务器 |
| 后端 HTTP Gateway | http://127.0.0.1:8888 | REST API 入口 |
| 后端 Kitex RPC | 127.0.0.1:8889 | Thrift RPC（LLM 凭证管理等） |

前端开发服务器已配置 `/api` 反向代理，会自动将 API 请求转发到 `http://127.0.0.1:8888`。

---

## 七、停止服务

```bash
# 如果 make dev 仍在前台运行，直接按 Ctrl+C

# 或强制清理端口
make stop
```

---

## 八、单独运行后端 / 前端

如需分别启动两个服务：

```bash
# 仅启动后端（会先停止旧进程）
make server

# 仅启动前端（新开一个终端）
make web
```

---

## 九、验证与测试

```bash
# 运行所有 Go 单元测试
go test ./...

# 静态代码检查
go vet ./...
```

---

## 十、常见问题

### 端口被占用

```bash
# 手动查看并终止占用进程
sudo lsof -ti :8888 | xargs kill -9
sudo lsof -ti :8889 | xargs kill -9
sudo lsof -ti :3000  | xargs kill -9

# 或直接用 Makefile 命令
make stop
```

### `LLM_GATEWAY_MASTER_KEY` 未设置或无效

启动时若看到类似错误：

```
FATAL: LLM_GATEWAY_MASTER_KEY missing or invalid
```

请检查：
1. 环境变量是否已 `export`（不能只写 `KEY=value`，必须带 `export`）
2. 密钥是否恰好 64 个十六进制字符（`echo ${#LLM_GATEWAY_MASTER_KEY}` 应输出 64）

### `go: command not found`

Go 的 `PATH` 未生效，执行：

```bash
source ~/.bashrc
# 或
export PATH=$PATH:/usr/local/go/bin
```

### `node: command not found` 或 npm 版本过低

通过 nvm 切换版本：

```bash
nvm use 18
```

或重新安装：

```bash
nvm install 18 && nvm use 18
```

### 前端依赖安装失败（npm ERR!）

尝试切换 npm 镜像源：

```bash
npm config set registry https://registry.npmmirror.com
cd web && npm install
```

---

## 附录：项目结构速览

```
blender-shot-preview/
├── cmd/server/          # 后端入口（Kitex + HTTP Gateway）
├── internal/
│   ├── gateway/         # HTTP API 层
│   ├── handler/         # RPC 处理器
│   ├── service/         # 业务逻辑（含 pipeline DAG）
│   └── agent/           # Agent 服务
├── idl/v0_1/            # Thrift IDL 定义
├── kitex_gen/           # Kitex 生成代码（勿手动修改）
├── web/                 # 前端 React + Vite + TypeScript
│   └── src/
├── docs/
│   ├── quickstart-linux.md          # 本文档
│   └── pipeline/shot-preview-runtime.md
├── Makefile
└── go.mod
```

---

> **说明**：当前版本使用**加密内存存储**保存 LLM 凭证，重启服务后凭证会丢失，需重新通过 RPC 写入。PostgreSQL 持久化存储为后续规划功能。
>
> 服务部署时请置于**可信内网或认证 RPC 边界之后**，不要直接暴露在公网。
