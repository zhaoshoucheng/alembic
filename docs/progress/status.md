# 当前进展与 Roadmap

> 最近更新：2026-06-08
> 模块层定义见 [module-map.md](./module-map.md)。✅ = 已落地，🟡 = 简版/部分，❌ = 未开始。

## 已完成对照表

| 层 | 状态 | 工程位置 | 说明 |
|---|---|---|---|
| A 配置 | ✅ | `infra/config/` | yaml + settings 加载 |
| A 账号/凭据 | ✅ | `domain/account/` | file + db 两种 loader，provider 抽象，凭据运行期查询 |
| A 可观测性 | ❌ | — | tracing / metrics 未做 |
| B 模型层 | ✅ | `domain/llm/` + `internal/llm/langchaingo/` | vendored langchaingo：openai / anthropic / gemini，含流式 |
| C 工具 registry | ✅ | `internal/tools/registry.go` | 线程安全注册表 + 工具定义导出 |
| C 内置工具 | 🟡 | `internal/tools/builtin/` | calculator / search / file / terminal / browser（search 偏 stub） |
| C 沙箱 | 🟡 | `internal/tools/sandbox/` | 本地沙箱已做，远程沙箱未做 |
| C 工具流式 | 🟡 | `internal/tools/streaming/` | router + json parser + 几个 stream handler |
| D 主循环 | 🟡 | `cmd/agent/agent_runtime.go` | **已跑通** function-calling 工具调用循环（非 ReAct 文本解析） |
| D 记忆 | ❌ | `internal/memory/`（空目录） | 每次 Run 都是全新消息，无跨轮上下文 |
| D 提示词 | 🟡 | `internal/prompt/templates.go` | 仅骨架模板 |
| D 计划/鲁棒性 | ❌ | — | 计划、循环检测、重试修复未做 |
| D 事件总线 | ❌ | — | 思考/进度事件未对上层分发 |
| E 会话与调度 | ❌ | — | WebSocket 会话、worker 分发未做 |
| F 服务端 | ❌ | — | HTTP/RPC、鉴权、路由未做 |

## 一句话现状

LLM、工具、单轮工具调用循环都通了。**最大的缺口是 Agent 没有记忆**，且提示词/计划/事件系统还很薄。

## Roadmap（候选下一步，按推荐顺序）

1. **记忆 Memory / 上下文窗口** — 在 `internal/memory/` 实现 BufferMemory（滑动窗口），
   接进 agent 循环，支持多轮对话。体量小、承接现有循环最顺。 ← *推荐先做*
2. **提示词系统 Prompt** — 完善 `internal/prompt/`，做系统提示词 + 模板渲染/解析。
3. **事件 / 流式输出** — 把 LLM 流式 + 工具进度通过事件总线吐给上层。
4. **会话与调度（E 层）** — WebSocket 会话 + worker 分发。
5. **服务端（F 层）** — HTTP/RPC 接口、鉴权。

> 下一步具体做哪个，由我（用户）后续再定。

## 维护约定

- 每完成一个模块，更新上面的对照表（状态 + 位置 + 说明）和"一句话现状"。
- 顶层 `CLAUDE.md` 的 Architecture 应与此表保持一致，发现不符就同步。
