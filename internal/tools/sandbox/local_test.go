package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalTerminal_Execute(t *testing.T) {
	sb := NewLocalSandbox(t.TempDir())
	ctx := context.Background()
	res, err := sb.Terminal().Execute(ctx, "s1", "echo hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(res.Stdout) != "hello" {
		t.Fatalf("got stdout=%q", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Fatalf("got exitCode=%d", res.ExitCode)
	}
}

func TestLocalTerminal_NonZeroExit(t *testing.T) {
	sb := NewLocalSandbox(t.TempDir())
	ctx := context.Background()
	res, err := sb.Terminal().Execute(ctx, "s1", "exit 42", "")
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestLocalTextEditor_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	sb := NewLocalSandbox(dir)
	ctx := context.Background()
	path := filepath.Join(dir, "test.txt")

	res, err := sb.TextEditor().RunAction(ctx, TextEditorCommand{
		Action:  ActionWrite,
		Path:    path,
		Content: "line1\nline2\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "ok" {
		t.Fatalf("write status=%s msg=%s", res.Status, res.Message)
	}

	res, err = sb.TextEditor().RunAction(ctx, TextEditorCommand{
		Action: ActionView,
		Path:   path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Content != "line1\nline2\n" {
		t.Fatalf("got content=%q", res.Content)
	}
}

func TestLocalTextEditor_Replace(t *testing.T) {
	dir := t.TempDir()
	sb := NewLocalSandbox(dir)
	ctx := context.Background()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	res, err := sb.TextEditor().RunAction(ctx, TextEditorCommand{
		Action: ActionReplace,
		Path:   path,
		OldStr: "world",
		NewStr: "Go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "ok" {
		t.Fatalf("replace status=%s", res.Status)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello Go" {
		t.Fatalf("got %q", string(data))
	}
}

func TestLocalFileSystem_CRUD(t *testing.T) {
	dir := t.TempDir()
	sb := NewLocalSandbox(dir)
	ctx := context.Background()
	fs := sb.FileSystem()
	path := filepath.Join(dir, "sub", "file.txt")

	exists, _ := fs.Exists(ctx, path)
	if exists {
		t.Fatal("file should not exist yet")
	}

	err := fs.WriteFile(ctx, path, "content")
	if err != nil {
		t.Fatal(err)
	}

	exists, _ = fs.Exists(ctx, path)
	if !exists {
		t.Fatal("file should exist after write")
	}

	content, err := fs.ReadFile(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if content != "content" {
		t.Fatalf("got content=%q", content)
	}

	entries, err := fs.ListDir(ctx, filepath.Join(dir, "sub"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0] != "file.txt" {
		t.Fatalf("got entries=%v", entries)
	}
}
