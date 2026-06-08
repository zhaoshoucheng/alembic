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

## Roadmap（按北极星目标"本地终端单文件 agent"排序）

> 方向以 [principles.md](./principles.md) 为准：只做 agent 本质（B/C/D），
> **E 会话调度 / F 服务端已明确砍掉**（属附带复杂度，对终端单文件目标无用）。

1. **记忆 Memory / 上下文窗口** — 在 `internal/memory/` 实现 BufferMemory（滑动窗口），
   接进 agent 循环，支持多轮对话。体量小、承接现有循环最顺。 ← *推荐先做*
2. **交互式终端 REPL** — 把"跑一次就退出"改成"输入→流式回答→再输入"的循环，
   这是"像 Claude 一样在终端交互"的关键一步。
3. **流式输出到终端** — 接上 LLM 层已有的流式能力，让回答边生成边显示。
4. **提示词系统 Prompt** — 完善 `internal/prompt/`，做系统提示词 + 模板渲染/解析。
5. **工具增强 + 权限确认** — 丰富 file/bash/search 等工具，跑高风险命令前先征求确认。
6. **鲁棒性** — 循环检测、重复调用检测、重试修复（参考工程 D 层有大量现成思路）。
7. **会话持久化 / resume** — 退出后能接着上次继续。

> ~~E 会话与调度 / F 服务端~~ — 不做，见 principles.md。
> 下一步具体做哪个，由我（用户）后续再定。

## 维护约定

- 每完成一个模块，更新上面的对照表（状态 + 位置 + 说明）和"一句话现状"。
- 顶层 `CLAUDE.md` 的 Architecture 应与此表保持一致，发现不符就同步。
