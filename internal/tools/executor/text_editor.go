package executor

import (
	"context"
	"fmt"

	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

type TextEditorExecutor struct {
	editor sandbox.TextEditor
}

func NewTextEditorExecutor(sb sandbox.Sandbox) *TextEditorExecutor {
	return &TextEditorExecutor{editor: sb.TextEditor()}
}

func (e *TextEditorExecutor) Name() string { return "text_editor" }

func (e *TextEditorExecutor) SupportedTools() []string {
	return []string{"file_read", "file_write_text", "file_replace_text", "file_append_text"}
}

func (e *TextEditorExecutor) Execute(ctx context.Context, toolName string, params map[string]any) Result {
	switch toolName {
	case "file_read":
		return e.fileRead(ctx, params)
	case "file_write_text":
		return e.fileWrite(ctx, params)
	case "file_replace_text":
		return e.fileReplace(ctx, params)
	case "file_append_text":
		return e.fileAppend(ctx, params)
	default:
		return Result{ToolName: toolName, Content: fmt.Sprintf("unsupported tool %q", toolName), IsError: true}
	}
}

func (e *TextEditorExecutor) fileRead(ctx context.Context, params map[string]any) Result {
	path, _ := params["path"].(string)
	if path == "" {
		path, _ = params["abs_path"].(string)
	}
	if path == "" {
		return Result{ToolName: "file_read", Content: "path is required", IsError: true}
	}

	res, err := e.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action: sandbox.ActionView,
		Path:   path,
	})
	if err != nil {
		return Result{ToolName: "file_read", Content: err.Error(), IsError: true}
	}
	if res.Status != "ok" {
		return Result{ToolName: "file_read", Content: res.Message, IsError: true}
	}
	return Result{ToolName: "file_read", Content: res.Content}
}

func (e *TextEditorExecutor) fileWrite(ctx context.Context, params map[string]any) Result {
	path, _ := params["path"].(string)
	if path == "" {
		path, _ = params["abs_path"].(string)
	}
	content, _ := params["file_text"].(string)
	if content == "" {
		content, _ = params["content"].(string)
	}
	if path == "" {
		return Result{ToolName: "file_write_text", Content: "path is required", IsError: true}
	}

	res, err := e.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action:  sandbox.ActionWrite,
		Path:    path,
		Content: content,
	})
	if err != nil {
		return Result{ToolName: "file_write_text", Content: err.Error(), IsError: true}
	}
	if res.Status != "ok" {
		return Result{ToolName: "file_write_text", Content: res.Message, IsError: true}
	}
	return Result{ToolName: "file_write_text", Content: res.Message}
}

func (e *TextEditorExecutor) fileReplace(ctx context.Context, params map[string]any) Result {
	path, _ := params["path"].(string)
	if path == "" {
		path, _ = params["abs_path"].(string)
	}
	oldStr, _ := params["old_str"].(string)
	newStr, _ := params["new_str"].(string)
	if path == "" {
		return Result{ToolName: "file_replace_text", Content: "path is required", IsError: true}
	}

	res, err := e.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action: sandbox.ActionReplace,
		Path:   path,
		OldStr: oldStr,
		NewStr: newStr,
	})
	if err != nil {
		return Result{ToolName: "file_replace_text", Content: err.Error(), IsError: true}
	}
	if res.Status != "ok" {
		return Result{ToolName: "file_replace_text", Content: res.Message, IsError: true}
	}
	return Result{ToolName: "file_replace_text", Content: res.Message}
}

func (e *TextEditorExecutor) fileAppend(ctx context.Context, params map[string]any) Result {
	path, _ := params["path"].(string)
	if path == "" {
		path, _ = params["abs_path"].(string)
	}
	content, _ := params["content"].(string)
	if content == "" {
		content, _ = params["file_text"].(string)
	}
	if path == "" {
		return Result{ToolName: "file_append_text", Content: "path is required", IsError: true}
	}
	line := 0
	if raw, ok := params["insert_line"]; ok {
		if f, ok := raw.(float64); ok {
			line = int(f)
		}
	}

	res, err := e.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action:  sandbox.ActionAppend,
		Path:    path,
		Content: content,
		Line:    line,
	})
	if err != nil {
		return Result{ToolName: "file_append_text", Content: err.Error(), IsError: true}
	}
	if res.Status != "ok" {
		return Result{ToolName: "file_append_text", Content: res.Message, IsError: true}
	}
	return Result{ToolName: "file_append_text", Content: res.Message}
}
