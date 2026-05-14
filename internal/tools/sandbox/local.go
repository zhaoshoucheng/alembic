package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LocalSandbox struct {
	workDir  string
	terminal *localTerminal
	editor   *localTextEditor
	fs       *localFileSystem
	browser  *localBrowser
}

func NewLocalSandbox(workDir string) *LocalSandbox {
	sb := &LocalSandbox{workDir: workDir}
	sb.terminal = &localTerminal{workDir: workDir}
	sb.editor = &localTextEditor{}
	sb.fs = &localFileSystem{}
	sb.browser = &localBrowser{}
	return sb
}

func (s *LocalSandbox) Terminal() Terminal     { return s.terminal }
func (s *LocalSandbox) TextEditor() TextEditor { return s.editor }
func (s *LocalSandbox) FileSystem() FileSystem { return s.fs }
func (s *LocalSandbox) Browser() Browser       { return s.browser }

// --- Terminal ---

type localTerminal struct {
	workDir string
}

func (t *localTerminal) Execute(ctx context.Context, _ string, command string, workingDir string) (*ExecResult, error) {
	dir := workingDir
	if dir == "" {
		dir = t.workDir
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	return &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

func (t *localTerminal) View(_ context.Context, _ string) (*ExecResult, error) {
	return &ExecResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil
}

// --- TextEditor ---

type localTextEditor struct{}

func (e *localTextEditor) RunAction(_ context.Context, cmd TextEditorCommand) (*TextEditorResult, error) {
	switch cmd.Action {
	case ActionView:
		data, err := os.ReadFile(cmd.Path)
		if err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		return &TextEditorResult{Status: "ok", Content: string(data)}, nil

	case ActionCreate, ActionWrite:
		dir := filepath.Dir(cmd.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		if err := os.WriteFile(cmd.Path, []byte(cmd.Content), 0o644); err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		return &TextEditorResult{Status: "ok", Message: "file written"}, nil

	case ActionReplace:
		data, err := os.ReadFile(cmd.Path)
		if err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		content := string(data)
		if !strings.Contains(content, cmd.OldStr) {
			return &TextEditorResult{Status: "error", Message: "old_str not found in file"}, nil
		}
		content = strings.Replace(content, cmd.OldStr, cmd.NewStr, 1)
		if err := os.WriteFile(cmd.Path, []byte(content), 0o644); err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		return &TextEditorResult{Status: "ok", Message: "replacement done"}, nil

	case ActionAppend:
		data, err := os.ReadFile(cmd.Path)
		if err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		lines := strings.Split(string(data), "\n")
		insertAt := cmd.Line
		if insertAt < 0 {
			insertAt = 0
		}
		if insertAt > len(lines) {
			insertAt = len(lines)
		}
		newLines := strings.Split(cmd.Content, "\n")
		result := make([]string, 0, len(lines)+len(newLines))
		result = append(result, lines[:insertAt]...)
		result = append(result, newLines...)
		result = append(result, lines[insertAt:]...)
		if err := os.WriteFile(cmd.Path, []byte(strings.Join(result, "\n")), 0o644); err != nil {
			return &TextEditorResult{Status: "error", Message: err.Error()}, nil
		}
		return &TextEditorResult{Status: "ok", Message: "text inserted"}, nil

	default:
		return &TextEditorResult{Status: "error", Message: fmt.Sprintf("unknown action %q", cmd.Action)}, nil
	}
}

func (e *localTextEditor) BatchRead(_ context.Context, paths []string) ([]FileInfo, error) {
	results := make([]FileInfo, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			results = append(results, FileInfo{Path: p, Exists: false})
			continue
		}
		results = append(results, FileInfo{Path: p, Content: string(data), Exists: true})
	}
	return results, nil
}

// --- FileSystem ---

type localFileSystem struct{}

func (f *localFileSystem) Exists(_ context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f *localFileSystem) ReadFile(_ context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *localFileSystem) WriteFile(_ context.Context, path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func (f *localFileSystem) ListDir(_ context.Context, path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	return names, nil
}

// --- Browser ---

type localBrowser struct{}

func (b *localBrowser) Execute(_ context.Context, action BrowserAction) (*BrowserResult, error) {
	return &BrowserResult{
		Content: fmt.Sprintf("browser action %q is not supported in local sandbox", action.ToolName),
		IsError: true,
		Error:   "local sandbox does not support browser operations",
	}, nil
}
