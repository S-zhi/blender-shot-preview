# Tool 执行检查单

本检查单适用于所有由 Agent 调用的 Skill / Tool，尤其是会访问外部服务、写数据、发消息或触发工作流的工具。

## 1. 统一执行上下文

每次工具调用在进入具体 Skill 前，必须生成并向下传递以下上下文：

```text
run_id             本次 Agent 运行
tool_call_id       本次模型工具调用的唯一 ID
idempotency_key    稳定幂等键，格式：agent:{run_id}:tool:{tool_call_id}
agent_id           调用的 Agent
tool_name          工具名称
tenant_id / user_id 调用身份
attempt            当前尝试次数
deadline           本次调用截止时间
```

`tool_call_id` 必须在第一次执行前确定，并在同一次调用的所有重试中保持不变。不能使用每次重试新生成的 UUID；否则重复请求无法被识别。

工具调用记录至少应保存：`run_id`、`tool_call_id`、`idempotency_key`、请求参数哈希、状态、尝试次数、最终结果或错误、开始和结束时间。

建议建立唯一约束：

```text
UNIQUE(run_id, tool_call_id)
UNIQUE(idempotency_key)
```

## 2. 参数和格式检查

模型产出的参数不可信。工具执行前必须完成以下检查，不满足即返回不可重试错误，不能调用下游业务服务。

| 检查项 | 要求 | 示例错误码 |
|---|---|---|
| JSON 格式 | 参数必须能解析为预期 JSON | `TOOL_INVALID_JSON` |
| Schema | 必填字段、类型、枚举、嵌套结构正确 | `TOOL_INVALID_ARGUMENT` |
| 未知字段 | 默认拒绝未声明字段，除非工具明确允许 | `TOOL_UNKNOWN_ARGUMENT` |
| 字符串限制 | 长度、字符集、敏感字符符合要求 | `TOOL_ARGUMENT_OUT_OF_RANGE` |
| 数值限制 | 金额、数量、分页大小、时间范围在业务上限内 | `TOOL_ARGUMENT_OUT_OF_RANGE` |
| 资源限制 | 单次文件数、请求体大小、返回结果大小受限 | `TOOL_LIMIT_EXCEEDED` |
| 业务前置条件 | 资源存在、状态允许、参数组合有效 | `TOOL_PRECONDITION_FAILED` |

典型“超值”错误包括：金额超过退款上限、分页 `page_size` 超过 100、批处理数量超过 500、检索时间范围超过 90 天、单次工具结果超过 256 KiB。上限必须由工具自己的业务规则定义，不能只依赖 Agent Prompt。

## 3. 幂等与重复执行

### 读工具

查询、检索、状态读取等无副作用工具可以自动重试。建议只对网络超时、连接失败、可恢复的 5xx 和限流错误重试，最多 3 次，使用指数退避和随机抖动。

### 写工具

退款、写数据库、发送消息、创建工单等工具必须支持幂等键。下游业务 Service 应按 `idempotency_key` 返回第一次成功的结果，而不是执行第二次。

遇到“请求已发出但响应超时”的不确定状态时：

```text
下游支持幂等键或可查询操作结果 -> 使用原幂等键安全重试或查询
下游不支持幂等键                 -> 标记 UNKNOWN，不自动重试，交给作业恢复或人工处理
```

同一幂等键配不同参数哈希必须返回 `TOOL_IDEMPOTENCY_CONFLICT`，不能悄悄复用旧结果。

模型在不同 `tool_call_id` 下连续发出相同写操作，不属于传输重试。应由业务 Service 的幂等规则和 Agent 的同工具次数限制共同拦截，不能仅靠重试机制处理。

## 4. 错误分类和返回格式

所有工具应转换为统一错误，不向模型或上层暴露数据库连接串、内部栈信息和供应商原始响应。

```text
VALIDATION       参数、格式或范围错误，不重试
AUTHORIZATION    权限或租户错误，不重试
NOT_FOUND        目标不存在，不重试
CONFLICT         状态冲突或并发修改，按业务决定是否重试
BUSINESS_RULE    违反业务规则，不重试
RATE_LIMITED     被限流，可按 retry_after 重试
TIMEOUT          调用超时；写操作先按幂等规则确认状态
UNAVAILABLE      网络或依赖服务不可用，可重试
INTERNAL         未分类内部错误，谨慎重试
LIMIT_EXCEEDED   超出大小、数量、额度或执行上限，不重试
UNKNOWN          副作用是否发生无法确认，不自动重试
```

推荐的错误载荷：

```json
{
  "code": "TOOL_ARGUMENT_OUT_OF_RANGE",
  "message": "page_size must be between 1 and 100",
  "retryable": false,
  "retry_after_ms": 0
}
```

`message` 可以给模型阅读，但必须是安全、可理解的描述；详细诊断写入受限日志和 trace。

## 5. 执行状态

工具调用状态建议为：

```text
PENDING -> RUNNING -> SUCCEEDED
                   -> FAILED
                   -> UNKNOWN
```

相同幂等键再次进入时的行为：

| 已有状态 | 处理方式 |
|---|---|
| `SUCCEEDED` | 参数哈希一致则返回已保存结果 |
| `RUNNING` | 不重复执行；等待、订阅结果或返回“执行中” |
| `FAILED` | 仅当错误可重试且未超过次数时，以原幂等键重试 |
| `UNKNOWN` | 不自动重试；先查询下游最终状态或进入人工/作业恢复 |

## 6. 必测场景

- 非 JSON 参数、缺少必填字段、字段类型错误、未知字段；
- 金额、数量、分页、时间范围和文件大小超过上限；
- 下游返回 4xx、5xx、限流、连接失败和超时；
- 读工具在第 2 或第 3 次重试成功；
- 写工具首次成功但响应超时，再调用只返回第一次结果；
- 相同幂等键但参数不同，返回冲突且不执行第二次；
- 两个并发请求使用相同幂等键，只产生一次副作用；
- 写工具出现不确定结果时状态为 `UNKNOWN`，没有自动二次写入；
- 工具返回不符合输出 Schema、结果过大或包含敏感字段；
- 取消、超时和服务重启后，调用记录仍能说明是否已经产生副作用。

## 7. 上线前确认

- 每个 Tool 标注为 `READ`、`WRITE` 或 `IRREVERSIBLE`；
- 每个 WRITE / IRREVERSIBLE Tool 的下游均已验证幂等键语义；
- 每个参数都有 Schema 和业务范围；
- 每个外部调用都有超时、错误映射和 trace；
- 自动重试仅覆盖明确可恢复的错误；
- 所有副作用操作都能通过 `run_id`、`tool_call_id` 和幂等键审计。
