package builtin

import (
	"github.com/shoucheng/my-first-agent/internal/tools"
	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

func NewBuiltinRegistry(sb sandbox.Sandbox) (*tools.Registry, error) {
	registry := tools.NewRegistry()
	ts := []tools.Tool{
		NewCalculator(),
		// terminal
		NewShellExec(sb),
		NewShellView(sb),
		// file
		NewFileRead(sb),
		NewFileWriteText(sb),
		NewFileReplaceText(sb),
		NewFileAppendText(sb),
		// browser
		NewBrowserNavigate(sb),
		NewBrowserView(sb),
		NewBrowserClick(sb),
		NewBrowserInput(sb),
		NewBrowserScrollUp(sb),
		NewBrowserScrollDown(sb),
		// search
		NewOmniSearch(),
	}
	for _, t := range ts {
		err := registry.Register(t)
		return registry, err
	}
	return registry, nil
}
