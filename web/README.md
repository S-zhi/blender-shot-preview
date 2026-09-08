# Blender Shot Preview - Web 前端

本文档专门介绍本前端工程的启动方式、已有功能内容以及前后端接口的实现/Mock 状态。

---

## 1. 如何快速启动

### 环境要求
- **Node.js**: `>= 18.0.0`
- **包管理器**: `npm`（或 `pnpm` / `yarn`）

### 启动步骤
1. **进入前端目录**：
   ```bash
   cd web
   ```
2. **安装依赖**：
   ```bash
   npm install
   ```
3. **启动本地开发服务器**：
   ```bash
   npm run dev
   ```
   启动后在浏览器打开终端提示的地址（默认 `http://localhost:3000`）。支持代码热更新 (HMR)。
4. **（可选）构建生产产物**：
   ```bash
   npm run build
   ```

---

## 2. 我们有哪些内容

系统深度参考了 **OpenHands (Agent Canvas)** 的交互规范与视觉设计，包含以下核心内容：

### 2.1 页面交互与功能
- **响应式侧边栏 (Sidebar)**：
  - “+ 新建分镜会话”快速创建独立会话；
  - 历史会话分组展示（按“今天”、“更早之前”归类）；
  - 实时会话关键词搜索与悬浮快速删除；
  - 侧边栏展开 / 极简迷你图标模式切换。
- **悬浮式卡片输入面板 (InputBox)**：
  - 自适应高度多行输入框（支持 `Enter` 发送、`Shift + Enter` 换行）；
  - 底部集成模型快捷指示器（显示当前提供商与模型名）；
  - 常用分镜运镜 Prompt 模版快速填充；
  - 任务生成中动态切换为停止按钮（可随时中断）。
- **可折叠思考链卡片 (Thought Process)**：
  - 仿 OpenHands 折叠组件，分步骤展示分镜语义分析、轨迹规划与渲染调度过程；
  - 关联展示 RPC 任务 ID 徽标；
  - 消息正文支持完整 Markdown 语法渲染。
- **多视图工作区 (Workspace Tabs)**：
  - **对话视图 (Chat)**：人机交互主界面，未输入时提供 4 组精选镜头场景推荐卡片；
  - **分镜视口 (Preview)**：预留 Blender 三维视口 / 渲染画面展示区；
  - **渲染日志 (Logs)**：展示后端服务与调度日志。
- **LLM Gateway 凭据管理弹窗 (Settings Modal)**：
  - 提供表单配置用户标识 (`user_id`)、服务商 (`LLMProvider`)、模型名 (`name`)、密钥 (`api_key`) 及网关地址 (`base_url`)。

### 2.2 代码与工程结构
- `src/components/`：侧边栏 (`sidebar/`)、对话核心区 (`conversation/`)、凭据弹窗 (`settings/`)、原子基础组件 (`ui/`)。
- `src/stores/`：Zustand 细粒度状态切片（`conversation-store.ts`, `sidebar-store.ts`, `settings-store.ts`）。
- `src/api/`：API 客户端封装、领域类型与服务门面。

---

## 3. 哪些接口已经实现，哪些接口是 mock 的

前端代码在 `src/api/` 目录下严格遵循与 UI 解耦的架构，状态划分如下：

### 3.1 已经实现的接口（契约完整 & 逻辑已打通）
| 接口 / 服务 | 对应后端契约 (IDL) | 实现状态 | 说明 |
| :--- | :--- | :--- | :--- |
| **`ShotPreviewService.createShotPreviewTask`** | `idl/v0_1/shot_preview.thrift`<br>`CreateShotPreviewTask` | **已实现** | 完整实现了 `user_id`, `prompt`, `conversation_id`, `request_id` 入参，及响应的 `task_id`、`TaskStatus` 处理。在 UI 上已打通用户输入、任务创建、思考卡片与 Task 徽标展示。 |
| **`LLMKeyService.manageLLMKey`** | `idl/v0_1/llm_gateway.thrift`<br>`ManageLLMKey` | **已实现** | 完整实现了 `user_id`, `operation` (SAVE/UPDATE/DELETE), `key_id`, `provider`, `name`, `api_key`, `base_url` 入参，及返回 `key_id`。在 UI 设置弹窗中已打通表单收集与保存状态。 |
| **会话状态与历史管理** | 前端本地状态流 | **已实现** | 会话的新建、切换、删除、标题动态提炼、过滤搜索全流程已完全打通。 |

### 3.2 当前处于 Mock 的部分及原因
| 模块 / 接口 | 当前状态 | 现状与原因 |
| :--- | :--- | :--- |
| **分镜任务实际后端调用** | **Mock** | 后端是基于 **Kitex Thrift RPC** 的内部服务，当前尚未架设面向浏览器的 HTTP Gateway 或 WebSocket 代理。因此前端通过 `src/api/mock-adapter.ts` 模拟网络耗时、返回模拟 `task_id`，并提供 OpenHands 风格的思考链流式演示。 |
| **LLM 凭据实际后端存储** | **Mock** | 同样因为后端是 Thrift RPC 协议，前端目前在 `mock-adapter.ts` 中模拟保存逻辑并即时返回 `key_id`。 |
| **Blender 画面实时视口 (Preview)** | **Mock / 占位** | 分镜视口目前为概念占位面板，待后续接入 WebRTC 实时推流、静态帧序列或 Three.js 视图。 |

### 3.3 如何从 Mock 切换为真实后端调用
1. 打开 [`src/api/client.ts`](src/api/client.ts)，将配置改为：
   ```typescript
   export const USE_MOCK_API = false;
   ```
2. 在 [`vite.config.ts`](vite.config.ts) 的 `server.proxy` 中配置后端的 HTTP 网关地址（默认已预留指向 `http://127.0.0.1:8888`）。
3. 前端发送的 `/api/v0_1/shot-preview/task` 和 `/api/v0_1/llm-gateway/key` 请求将自动透明转发至真实服务。
