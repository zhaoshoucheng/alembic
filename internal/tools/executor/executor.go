package executor

import (
	"context"
	"fmt"

	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

type Result struct {
	ToolName string
	Content  string
	IsError  bool
}

type Executor interface {
	Name() string
	SupportedTools() []string
	Execute(ctx context.Context, toolName string, params map[string]any) Result
}

type Manager struct {
	executors map[string]Executor
	sandbox   sandbox.Sandbox
}

func NewManager(sb sandbox.Sandbox) *Manager {
	return &Manager{
		executors: make(map[string]Executor),
		sandbox:   sb,
	}
}

func (m *Manager) Register(executor Executor) {
	for _, name := range executor.SupportedTools() {
		m.executors[name] = executor
	}
}

func (m *Manager) Execute(ctx context.Context, toolName string, params map[string]any) Result {
	executor, ok := m.executors[toolName]
	if !ok {
		return Result{
			ToolName: toolName,
			Content:  fmt.Sprintf("no executor registered for tool %q", toolName),
			IsError:  true,
		}
	}
	return executor.Execute(ctx, toolName, params)
}

func (m *Manager) HasExecutor(toolName string) bool {
	_, ok := m.executors[toolName]
	return ok
}

func (m *Manager) Sandbox() sandbox.Sandbox {
	return m.sandbox
}
