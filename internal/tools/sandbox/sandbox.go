package sandbox

import "context"

type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type FileInfo struct {
	Path    string
	Content string
	Exists  bool
}

type Terminal interface {
	Execute(ctx context.Context, sessionID string, command string, workingDir string) (*ExecResult, error)
	View(ctx context.Context, sessionID string) (*ExecResult, error)
}

type TextEditorAction string

const (
	ActionView    TextEditorAction = "view"
	ActionCreate  TextEditorAction = "create"
	ActionWrite   TextEditorAction = "write"
	ActionReplace TextEditorAction = "str_replace"
	ActionAppend  TextEditorAction = "insert"
)

type TextEditorCommand struct {
	Action  TextEditorAction
	Path    string
	Content string
	OldStr  string
	NewStr  string
	Line    int
}

type TextEditorResult struct {
	Status  string
	Content string
	Message string
}

type TextEditor interface {
	RunAction(ctx context.Context, cmd TextEditorCommand) (*TextEditorResult, error)
	BatchRead(ctx context.Context, paths []string) ([]FileInfo, error)
}

type FileSystem interface {
	Exists(ctx context.Context, path string) (bool, error)
	ReadFile(ctx context.Context, path string) (string, error)
	WriteFile(ctx context.Context, path string, content string) error
	ListDir(ctx context.Context, path string) ([]string, error)
}

type BrowserAction struct {
	ToolName string
	Params   map[string]any
}

type BrowserResult struct {
	URL        string
	Title      string
	Content    string
	Screenshot string
	Elements   []map[string]any
	IsError    bool
	Error      string
}

type Browser interface {
	Execute(ctx context.Context, action BrowserAction) (*BrowserResult, error)
}

type Sandbox interface {
	Terminal() Terminal
	TextEditor() TextEditor
	FileSystem() FileSystem
	Browser() Browser
}
