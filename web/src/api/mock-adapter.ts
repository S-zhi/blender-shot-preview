import {
  CreateShotPreviewTaskRequest,
  CreateShotPreviewTaskResponse,
  TaskStatus,
  ManageLLMKeyRequest,
  ManageLLMKeyResponse,
} from "./types";

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export const mockAdapter = {
  async createShotPreviewTask(
    req: CreateShotPreviewTaskRequest
  ): Promise<CreateShotPreviewTaskResponse> {
    await sleep(400);
    const taskId = `task_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`;
    return {
      task_id: taskId,
      status: TaskStatus.ACCEPTED,
      request_id: req.request_id || `req_${Date.now()}`,
    };
  },

  async manageLLMKey(
    req: ManageLLMKeyRequest
  ): Promise<ManageLLMKeyResponse> {
    await sleep(300);
    const keyId = req.key_id || `key_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`;
    return {
      key_id: keyId,
      provider: req.provider,
    };
  },

  /**
   * Helper to simulate OpenHands-like step-by-step thinking for a prompt
   */
  async simulateAgentStream(
    prompt: string,
    onThought: (thought: string) => void,
    onContent: (chunk: string) => void
  ) {
    onThought("正在解析分镜提示词并分析 Blender 场景上下文...");
    await sleep(600);

    onThought("构建镜头运动轨迹 (Camera Orbit & Pan 曲线计算)...");
    await sleep(800);

    onThought("检查灯光布局与材质渲染参数，准备调度 Blender 后台渲染...");
    await sleep(600);

    const fullResponse = `### 分镜预览任务已生成并排队

已根据您的需求解析分镜脚本：
- **镜头描述**：${prompt}
- **运镜模式**：特写平移至中景环绕 (Pan & Orbit)
- **渲染引擎**：Eevee Next (低延时实时预览)
- **帧率 / 范围**：24 FPS (1 ~ 120 帧)

> 💡 **提示**：后台渲染节点正在根据当前模型配置进行批处理生成，完成后的镜头序列帧将同步输出至面板。`;

    onContent(fullResponse);
  },
};
