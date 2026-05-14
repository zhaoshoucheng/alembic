package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

type TerminalExecutor struct {
	terminal sandbox.Terminal
}

func NewTerminalExecutor(sb sandbox.Sandbox) *TerminalExecutor {
	return &TerminalExecutor{terminal: sb.Terminal()}
}

func (e *TerminalExecutor) Name() string { return "terminal" }

func (e *TerminalExecutor) SupportedTools() []string {
	return []string{"shell_exec", "shell_view"}
}

func (e *TerminalExecutor) Execute(ctx context.Context, toolName string, params map[string]any) Result {
	switch toolName {
	case "shell_exec":
		return e.execShell(ctx, params)
	case "shell_view":
		return e.viewShell(ctx, params)
	default:
		return Result{ToolName: toolName, Content: fmt.Sprintf("unsupported tool %q", toolName), IsError: true}
	}
}

func (e *TerminalExecutor) execShell(ctx context.Context, params map[string]any) Result {
	command, _ := params["command"].(string)
	if command == "" {
		return Result{ToolName: "shell_exec", Content: "command is required", IsError: true}
	}
	sessionID, _ := params["session_id"].(string)
	workDir, _ := params["working_dir"].(string)

	res, err := e.terminal.Execute(ctx, sessionID, command, workDir)
	if err != nil {
		return Result{ToolName: "shell_exec", Content: err.Error(), IsError: true}
	}

	var output strings.Builder
	if res.Stdout != "" {
		output.WriteString(res.Stdout)
	}
	if res.Stderr != "" {
		if output.Len() > 0 {
			output.WriteByte('\n')
		}
		output.WriteString("STDERR: ")
		output.WriteString(res.Stderr)
	}
	if output.Len() == 0 {
		output.WriteString(fmt.Sprintf("Command completed with exit code %d", res.ExitCode))
	}

	return Result{
		ToolName: "shell_exec",
		Content:  output.String(),
		IsError:  res.ExitCode != 0,
	}
}

func (e *TerminalExecutor) viewShell(ctx context.Context, params map[string]any) Result {
	sessionID, _ := params["session_id"].(string)
	res, err := e.terminal.View(ctx, sessionID)
	if err != nil {
		return Result{ToolName: "shell_view", Content: err.Error(), IsError: true}
	}
	content := res.Stdout
	if content == "" {
		content = "No output available."
	}
	return Result{ToolName: "shell_view", Content: content}
}
