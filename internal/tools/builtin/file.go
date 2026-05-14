package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/shoucheng/my-first-agent/internal/tools"
	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

func filePath(args map[string]any) string {
	if p, ok := args["abs_path"].(string); ok && p != "" {
		return p
	}
	p, _ := args["path"].(string)
	return p
}

// --- file_read ---

type FileRead struct {
	editor sandbox.TextEditor
}

func NewFileRead(sb sandbox.Sandbox) *FileRead {
	return &FileRead{editor: sb.TextEditor()}
}

func (t *FileRead) Name() string { return "file_read" }

func (t *FileRead) Description() string {
	return tools.NormalizeDescription(
		"Read file content.",
		"When to use: reading text files, viewing images, reading PDFs/Word/PowerPoint as text or images.",
		"Best practices: prefer this tool over shell commands for file reading. Use line range limits appropriately.",
	)
}

func (t *FileRead) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":    tools.StringProperty("Brief one-sentence status description"),
		"abs_path": tools.StringProperty("Absolute path of the file to read"),
		"view_type": map[string]any{
			"type":        "string",
			"description": "Type of content to read: 'text' for text files, 'image' for image files",
			"enum":        []any{"text", "image"},
		},
		"line_range": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "integer"},
			"minItems":    2,
			"maxItems":    2,
			"description": "Line range [start, end). 0-based. Negative values count from end.",
		},
	}, "brief", "abs_path", "view_type")
}

func (t *FileRead) Execute(ctx context.Context, args map[string]any) (string, error) {
	path := filePath(args)
	if path == "" {
		return "", fmt.Errorf("abs_path is required")
	}

	res, err := t.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action: sandbox.ActionView,
		Path:   path,
	})
	if err != nil {
		return "", err
	}
	if res.Status != "ok" {
		return "", fmt.Errorf("%s", res.Message)
	}

	content := res.Content
	if lineRange, ok := args["line_range"].([]any); ok && len(lineRange) == 2 {
		content = applyLineRange(content, lineRange)
	}
	return content, nil
}

func applyLineRange(content string, lineRange []any) string {
	lines := strings.Split(content, "\n")
	total := len(lines)

	start := toInt(lineRange[0])
	end := toInt(lineRange[1])

	if start < 0 {
		start = total + start
	}
	if end < 0 {
		end = total + end
	}
	if start < 0 {
		start = 0
	}
	if end > total {
		end = total
	}
	if start >= end {
		return ""
	}
	return strings.Join(lines[start:end], "\n")
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// --- file_write_text ---

type FileWriteText struct {
	editor sandbox.TextEditor
}

func NewFileWriteText(sb sandbox.Sandbox) *FileWriteText {
	return &FileWriteText{editor: sb.TextEditor()}
}

func (t *FileWriteText) Name() string { return "file_write_text" }

func (t *FileWriteText) Description() string {
	return tools.NormalizeDescription(
		"Overwrite the content of a text file.",
		"When to use: creating new text files, overwriting existing files with completely new content.",
		"Best practices: prefer file_replace_text for updating specific content. DO NOT output snipped or truncated content.",
	)
}

func (t *FileWriteText) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":    tools.StringProperty("One brief sentence to explain this action"),
		"abs_path": tools.StringProperty("Absolute path of the file to write to"),
		"content":  tools.StringProperty("Text content to write"),
		"append_newline": map[string]any{
			"type":        "boolean",
			"description": "Whether to append a newline at the end. Defaults to true.",
		},
	}, "brief", "abs_path", "content")
}

func (t *FileWriteText) Execute(ctx context.Context, args map[string]any) (string, error) {
	path := filePath(args)
	if path == "" {
		return "", fmt.Errorf("abs_path is required")
	}
	content, _ := args["content"].(string)

	appendNewline := true
	if v, ok := args["append_newline"].(bool); ok {
		appendNewline = v
	}
	if appendNewline && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	res, err := t.editor.RunAction(ctx, sandbox.TextEditorCommand{
		Action:  sandbox.ActionWrite,
		Path:    path,
		Content: content,
	})
	if err != nil {
		return "", err
	}
	if res.Status != "ok" {
		return "", fmt.Errorf("%s", res.Message)
	}
	return fmt.Sprintf("File written: %s", path), nil
}

// --- file_replace_text ---

type FileReplaceText struct {
	editor sandbox.TextEditor
	fs     sandbox.FileSystem
}

func NewFileReplaceText(sb sandbox.Sandbox) *FileReplaceText {
	return &FileReplaceText{editor: sb.TextEditor(), fs: sb.FileSystem()}
}

func (t *FileReplaceText) Name() string { return "file_replace_text" }

func (t *FileReplaceText) Description() string {
	return tools.NormalizeDescription(
		"Replace specified string in a text file.",
		"When to use: updating specific content in files, fixing errors in code files.",
		"Best practices: old_str must exactly match content in the file. Use replace_all for multiple occurrences.",
	)
}

func (t *FileReplaceText) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":    tools.StringProperty("Brief one-sentence status description"),
		"abs_path": tools.StringProperty("Absolute path of the file"),
		"old_str":  tools.StringProperty("The exact string to be replaced. Must exist and appear only once unless replace_all is true."),
		"new_str":  tools.StringProperty("The string to replace old_str with"),
		"replace_all": map[string]any{
			"type":        "boolean",
			"description": "Whether to replace all occurrences. Defaults to false.",
		},
	}, "brief", "abs_path", "old_str", "new_str")
}

func (t *FileReplaceText) Execute(ctx context.Context, args map[string]any) (string, error) {
	path := filePath(args)
	if path == "" {
		return "", fmt.Errorf("abs_path is required")
	}
	oldStr, _ := args["old_str"].(string)
	newStr, _ := args["new_str"].(string)

	replaceAll := false
	if v, ok := args["replace_all"].(bool); ok {
		replaceAll = v
	}

	content, err := t.fs.ReadFile(ctx, path)
	if err != nil {
		return "", err
	}

	count := strings.Count(content, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_str not found in file")
	}
	if count > 1 && !replaceAll {
		return "", fmt.Errorf("old_str appears %d times in file; set replace_all=true to replace all", count)
	}

	n := 1
	if replaceAll {
		n = -1
	}
	content = strings.Replace(content, oldStr, newStr, n)

	if err := t.fs.WriteFile(ctx, path, content); err != nil {
		return "", err
	}
	return fmt.Sprintf("Replaced in %s (%d occurrence(s))", path, count), nil
}

// --- file_append_text ---

type FileAppendText struct {
	fs sandbox.FileSystem
}

func NewFileAppendText(sb sandbox.Sandbox) *FileAppendText {
	return &FileAppendText{fs: sb.FileSystem()}
}

func (t *FileAppendText) Name() string { return "file_append_text" }

func (t *FileAppendText) Description() string {
	return tools.NormalizeDescription(
		"Append content to a text file.",
		"When to use: adding new content to an existing file without overwriting it.",
		"Best practices: DO NOT output snipped or truncated content, always output full content.",
	)
}

func (t *FileAppendText) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":    tools.StringProperty("Brief one-sentence status description"),
		"abs_path": tools.StringProperty("Absolute path of the file to append to"),
		"content":  tools.StringProperty("Text content to append"),
		"append_newline": map[string]any{
			"type":        "boolean",
			"description": "Whether to append a newline at the end. Defaults to true.",
		},
	}, "brief", "abs_path", "content")
}

func (t *FileAppendText) Execute(ctx context.Context, args map[string]any) (string, error) {
	path := filePath(args)
	if path == "" {
		return "", fmt.Errorf("abs_path is required")
	}
	content, _ := args["content"].(string)

	appendNewline := true
	if v, ok := args["append_newline"].(bool); ok {
		appendNewline = v
	}
	if appendNewline && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	existing, err := t.fs.ReadFile(ctx, path)
	if err != nil {
		existing = ""
	}

	if err := t.fs.WriteFile(ctx, path, existing+content); err != nil {
		return "", err
	}
	return fmt.Sprintf("Content appended to %s", path), nil
}
