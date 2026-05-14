package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/shoucheng/my-first-agent/internal/tools"
	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

// --- shell_exec ---

type ShellExec struct {
	terminal sandbox.Terminal
}

func NewShellExec(sb sandbox.Sandbox) *ShellExec {
	return &ShellExec{terminal: sb.Terminal()}
}

func (t *ShellExec) Name() string { return "shell_exec" }

func (t *ShellExec) Description() string {
	return tools.NormalizeDescription(
		"Execute command in a shell session.",
		"When to use: running shell commands or scripts, installing packages, copying/moving/deleting files.",
		"Best practices: DO NOT use this tool to read and write files; use the file tools instead.",
	)
}

func (t *ShellExec) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":       tools.StringProperty("One brief sentence to explain this action"),
		"session_id":  tools.StringProperty("Unique identifier of the target shell session; automatically creates new session if not exists"),
		"working_dir": tools.StringProperty("Absolute path to the working directory for command execution"),
		"command":     tools.StringProperty("Shell command to execute"),
	}, "brief", "session_id", "working_dir", "command")
}

func (t *ShellExec) Execute(ctx context.Context, args map[string]any) (string, error) {
	command, _ := args["command"].(string)
	if command == "" {
		return "", fmt.Errorf("command is required")
	}
	sessionID, _ := args["session_id"].(string)
	workDir, _ := args["working_dir"].(string)

	res, err := t.terminal.Execute(ctx, sessionID, command, workDir)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	if res.Stdout != "" {
		out.WriteString(res.Stdout)
	}
	if res.Stderr != "" {
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString("STDERR: ")
		out.WriteString(res.Stderr)
	}
	if out.Len() == 0 {
		out.WriteString(fmt.Sprintf("Command completed with exit code %d", res.ExitCode))
	}
	if res.ExitCode != 0 {
		return out.String(), fmt.Errorf("exit code %d: %s", res.ExitCode, out.String())
	}
	return out.String(), nil
}

// --- shell_view ---

type ShellView struct {
	terminal sandbox.Terminal
}

func NewShellView(sb sandbox.Sandbox) *ShellView {
	return &ShellView{terminal: sb.Terminal()}
}

func (t *ShellView) Name() string { return "shell_view" }

func (t *ShellView) Description() string {
	return tools.NormalizeDescription(
		"View the content of a shell session.",
		"When to use: checking shell session history and current status, monitoring output of long-running processes.",
	)
}

func (t *ShellView) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":      tools.StringProperty("One brief sentence to explain this action"),
		"session_id": tools.StringProperty("Unique identifier of the shell session to view"),
	}, "brief", "session_id")
}

func (t *ShellView) Execute(ctx context.Context, args map[string]any) (string, error) {
	sessionID, _ := args["session_id"].(string)
	res, err := t.terminal.View(ctx, sessionID)
	if err != nil {
		return "", err
	}
	if res.Stdout == "" {
		return "No output available.", nil
	}
	return res.Stdout, nil
}
