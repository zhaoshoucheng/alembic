# Tools 模块设计文档

## 概述

`internal/tools/` 是 agent 的工具管理模块，负责工具的定义、注册、参数校验、执行和 LLM 流式响应处理。采用分层设计，将"工具是什么"和"工具怎么执行"解耦，并通过 Sandbox 抽象层实现执行环境的可迁移性。

## 架构总览

```
                          LLM
                           │
                    ┌──────┴──────┐
                    │  streaming  │  流式解析 LLM 输出中的 tool_call
                    │  Router     │  增量 JSON 解析 → StreamHandler 生命周期回调
                    └──────┬──────┘
                           │ 完整的 tool_call
                           ▼
                    ┌─────────────┐
                    │  Registry   │  工具注册表：查找、校验、分发
                    └──────┬──────┘
                           │ 校验通过的 args
                           ▼
                    ┌─────────────┐
                    │  builtin/*  │  具体工具实现（Tool 接口）
                    └──────┬──────┘
                           │ 调用 sandbox 接口
                           ▼
                    ┌─────────────┐
                    │  sandbox    │  执行环境抽象（Terminal/TextEditor/FileSystem/Browser）
                    │  ├ local    │  本地实现（os/exec, os.ReadFile ...）
                    │  └ remote?  │  远程实现（HTTP/WebSocket → 容器，未来扩展）
                    └─────────────┘
```

## 目录结构

```
internal/tools/
├── tool.go                 # Tool 接口 + JSONSchema 辅助函数
├── registry.go             # Registry：注册、查找、参数解析/校验、执行、结果序列化
│
├── builtin/                # 所有内置工具实现
│   ├── register.go         # NewBuiltinRegistry —— 按配置文件中的名字列表注册工具
│   ├── calculator.go       # calculator —— 算术表达式求值
│   ├── terminal.go         # shell_exec, shell_view —— 终端命令执行
│   ├── file.go             # file_read, file_write_text, file_replace_text, file_append_text
│   ├── browser.go          # browser_navigate/view/click/input/scroll_up/scroll_down
│   └── search.go           # omni_search —— 统一搜索
│
├── streaming/              # LLM 流式输出处理
│   ├── json_parser.go      # 增量 JSON 解析器（自动补全不完整的 JSON）
│   ├── handler.go          # StreamHandler 接口 + 事件类型
│   ├── base_handler.go     # BaseStreamHandler（公共状态管理）
│   ├── router.go           # StreamingRouter —— 按工具名分发 chunk 到 handler
│   ├── terminal_stream_handler.go
│   ├── text_editor_stream_handler.go
│   └── search_stream_handler.go
│
├── sandbox/                # 执行环境抽象层
│   ├── sandbox.go          # 接口定义：Terminal, TextEditor, FileSystem, Browser
│   └── local.go            # LocalSandbox —— 本地执行实现
│
└── executor/               # 执行器层（按工具类别聚合的执行逻辑）
    ├── executor.go         # Executor 接口 + Manager
    ├── terminal.go         # TerminalExecutor
    └── text_editor.go      # TextEditorExecutor
```

## 核心接口

### Tool 接口（tool.go）

每个工具必须实现的四个方法：

```go
type Tool interface {
    Name() string                                           // 工具名，如 "shell_exec"
    Description() string                                    // 发给 LLM 的描述
    Parameters() JSONSchema                                 // 参数的 JSON Schema
    Execute(ctx context.Context, args map[string]any) (string, error)  // 执行逻辑
}
```

- `Name()` + `Description()` + `Parameters()` 构成工具的**声明**，通过 `Definition()` 转换为 OpenAI function calling 格式发给 LLM。
- `Execute()` 是工具的**实现**，包含实际的业务逻辑（调用 sandbox、发 HTTP 请求等）。

### Registry（registry.go）

线程安全的工具注册表，提供完整的工具调用生命周期：

| 方法 | 作用 |
|------|------|
| `Register(tool)` | 注册工具（重名报错） |
| `Get(name)` | 按名查找 |
| `Definitions()` | 生成所有工具的 LLM function 定义（按名字排序，保证稳定） |
| `ParseToolCall(call)` | 解析 LLM 返回的 tool_call：查找工具 → JSON 解析（带容错修复）→ schema 校验 |
| `ExecuteToolCall(ctx, call)` | 解析 + 执行，返回 `CallResult` |
| `ExecuteToolCalls(ctx, calls)` | 批量执行 |
| `ResultMessage(result)` | 将 `CallResult` 序列化为 LLM 的 tool message 格式 |

解析结果是三态的（`ParseStatus`）：
- `ok` — 工具存在且参数校验通过
- `unknown_tool` — 工具名不在注册表中
- `wrong_args` — 工具存在但参数不合法

参数解析内置 JSON 容错修复（`json-repair`），能处理 LLM 偶尔返回的格式不规范的 JSON。

## 各层详解

### builtin/ — 工具实现

每个文件对应一类工具，一个 struct 对应一个具体工具。工具分为两种：

1. **需要 Sandbox 的工具**（shell、file、browser）：构造时注入 `sandbox.Sandbox`，Execute 中调用 sandbox 接口。
2. **自包含的工具**（calculator、omni_search）：不依赖 sandbox，自行完成计算或 HTTP 调用。

#### 当前内置工具一览

| 类别 | 工具名 | 文件 | 依赖 |
|------|--------|------|------|
| 计算 | `calculator` | calculator.go | 无 |
| 终端 | `shell_exec` | terminal.go | sandbox.Terminal |
| 终端 | `shell_view` | terminal.go | sandbox.Terminal |
| 文件 | `file_read` | file.go | sandbox.TextEditor |
| 文件 | `file_write_text` | file.go | sandbox.TextEditor |
| 文件 | `file_replace_text` | file.go | sandbox.TextEditor + FileSystem |
| 文件 | `file_append_text` | file.go | sandbox.FileSystem |
| 浏览器 | `browser_navigate` | browser.go | sandbox.Browser |
| 浏览器 | `browser_view` | browser.go | sandbox.Browser |
| 浏览器 | `browser_click` | browser.go | sandbox.Browser |
| 浏览器 | `browser_input` | browser.go | sandbox.Browser |
| 浏览器 | `browser_scroll_up` | browser.go | sandbox.Browser |
| 浏览器 | `browser_scroll_down` | browser.go | sandbox.Browser |
| 搜索 | `omni_search` | search.go | 无（直接 HTTP） |

#### 注册（register.go）

`NewBuiltinRegistry(settings, sandbox)` 读取 `config.yaml` 中的 `tools.enabled` 列表，按名字通过 switch-case 实例化对应工具并注册。未配置时默认只启用 `calculator`。

```yaml
# config.yaml
tools:
  enabled:
    - calculator
    - shell_exec
    - shell_view
    - file_read
    - file_write_text
    - file_replace_text
    - omni_search
```

### streaming/ — 流式处理

处理 LLM 流式输出中的 tool_call，在工具参数还没完全生成时就能实时推送进度。

#### 数据流

```
LLM chunk (含 tool_call 片段)
    │
    ▼
StreamingRouter.ProcessChunk(DeltaToolCall{ID, Name, ArgChunk})
    │
    ├─ 检测到新工具 → reset JsonStreamParser → handler.OnToolStart()
    │
    ├─ ArgChunk → JsonStreamParser.Parse()
    │              ├─ 拼接到缓冲区
    │              ├─ repairJSON() 补全引号/括号
    │              ├─ JSON 反序列化
    │              └─ 与上次结果比较 → changed?
    │                  └─ yes → handler.OnParamsDelta(partialParams)
    │
    └─ LLM 完成 → router.FinishCurrentTool() → handler.OnToolFinish(finalParams)
```

#### StreamHandler 接口

每个工具类别一个 handler，实现四个生命周期钩子：

```go
type StreamHandler interface {
    ToolName() string
    OnToolStart(toolID string, pusher StreamEventPusher) error
    OnParamsDelta(currentParams map[string]any, pusher StreamEventPusher) error
    OnToolFinish(finalParams map[string]any, pusher StreamEventPusher) (*FinishedToolInfo, error)
    OnToolAbort(reason string, pusher StreamEventPusher) error
}
```

通过 `StreamEventPusher` 接口将 `ToolDeltaEvent` 推送给上层（前端 UI 或日志系统），事件包含进度状态（init → delta → argumentsFinished → done / rollback）。

#### 增量 JSON 解析器（json_parser.go）

核心算法：跟踪字符串状态和括号栈，对不完整的 JSON 片段自动补全：

- `{"command": "ls -l` → 补 `"` 和 `}` → 解析为 `{"command": "ls -l"}`
- `{"a": 1, "b":` → 去除 trailing `,` / `:` → 补 `}` → 解析为 `{"a": 1}`

### sandbox/ — 执行环境抽象

将工具的执行逻辑与具体的运行环境解耦。

```go
type Sandbox interface {
    Terminal() Terminal       // 命令执行
    TextEditor() TextEditor   // 文件读写
    FileSystem() FileSystem   // 文件系统操作
    Browser() Browser         // 浏览器自动化
}
```

#### 接口设计原则

接口按操作类别划分（而非按工具划分），对齐 sandbox runtime API 端点设计：

| 接口 | 对应 sandbox endpoint | 主要方法 |
|------|---------------------|---------|
| Terminal | WebSocket `/ws/terminal` | Execute, View |
| TextEditor | `POST /text_editor` | RunAction, BatchRead |
| FileSystem | `POST /file/*` | Exists, ReadFile, WriteFile, ListDir |
| Browser | `POST /browser/action` | Execute |

#### 实现

- **LocalSandbox**（local.go）：当前唯一实现。Terminal 通过 `os/exec` 执行 shell 命令，TextEditor/FileSystem 通过 `os` 包直接读写本地文件，Browser 返回 "不支持" 的占位响应。
- **未来扩展**：实现 `RemoteSandbox`，通过 HTTP/WebSocket 调用远程容器中的 sandbox runtime，即可在不修改任何工具代码的情况下切换到容器化执行。

### executor/ — 执行器层

按工具类别聚合的执行逻辑，提供比单个 Tool 更高层的抽象：

```go
type Executor interface {
    Name() string
    SupportedTools() []string
    Execute(ctx context.Context, toolName string, params map[string]any) Result
}
```

`Manager` 维护 toolName → Executor 的映射，支持按工具名路由执行。当需要在 builtin 工具之外增加执行前/后的通用逻辑（日志、重试、限流）时，可以在 executor 层统一处理。

## 新增工具的步骤

1. 在 `builtin/` 下新建或编辑文件，实现 `tools.Tool` 接口：

```go
type MyTool struct {
    terminal sandbox.Terminal  // 按需注入 sandbox 组件
}

func NewMyTool(sb sandbox.Sandbox) *MyTool {
    return &MyTool{terminal: sb.Terminal()}
}

func (t *MyTool) Name() string        { return "my_tool" }
func (t *MyTool) Description() string { return "..." }
func (t *MyTool) Parameters() tools.JSONSchema {
    return tools.ObjectSchema(map[string]any{
        "param1": tools.StringProperty("description"),
    }, "param1")
}
func (t *MyTool) Execute(ctx context.Context, args map[string]any) (string, error) {
    // 通过 sandbox 执行
}
```

2. 在 `builtin/register.go` 的 switch 中添加一行：

```go
case "my_tool":
    tool = NewMyTool(sb)
```

3. 在 `config.yaml` 的 `tools.enabled` 中加上 `my_tool`。

4. （可选）如果需要流式进度推送，在 `streaming/` 下添加对应的 `StreamHandler` 并注册到 `StreamingRouter`。

## 设计决策

| 决策 | 理由 |
|------|------|
| 硬编码注册而非 YAML 动态加载 | 当前工具数量有限（~15 个），硬编码更简单、类型安全 |
| Tool 接口合并声明与实现 | 相比 YAML 声明 + Executor 分离的模式，Go 单接口更简洁直观 |
| Sandbox 接口按操作类别划分 | 对齐 sandbox runtime API 端点，便于未来迁移到远程容器 |
| 同步执行 | 当前版本足够；后续可用 goroutine 实现异步，无需改接口 |
| 增量 JSON 解析器自研 | 依赖轻量，核心算法仅括号/引号追踪 + 补全，~100 行代码 |
