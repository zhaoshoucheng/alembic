package builtin

import (
	"context"
	"fmt"

	"github.com/shoucheng/my-first-agent/internal/tools"
	"github.com/shoucheng/my-first-agent/internal/tools/sandbox"
)

// --- browser_navigate ---

type BrowserNavigate struct {
	browser sandbox.Browser
}

func NewBrowserNavigate(sb sandbox.Sandbox) *BrowserNavigate {
	return &BrowserNavigate{browser: sb.Browser()}
}

func (t *BrowserNavigate) Name() string { return "browser_navigate" }

func (t *BrowserNavigate) Description() string {
	return tools.NormalizeDescription(
		"Navigate the browser to a specified URL.",
		"When to use: visiting a specific web page, following search result links, refreshing current page.",
		"Best practices: check page response status. URL must include protocol prefix (e.g., https://).",
	)
}

func (t *BrowserNavigate) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief": tools.StringProperty("One brief sentence to explain this action"),
		"url":   tools.StringProperty("URL to navigate to. Must include protocol prefix (e.g., https://)"),
		"intent": map[string]any{
			"type":        "string",
			"description": "Purpose of navigation: navigational (general browsing), informational (reading content), transactional (performing actions)",
			"enum":        []any{"navigational", "informational", "transactional"},
		},
		"focus": tools.StringProperty("(Required if intent is informational) Specific topic or question to focus on"),
	}, "brief", "url", "intent")
}

func (t *BrowserNavigate) Execute(ctx context.Context, args map[string]any) (string, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_navigate",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- browser_view ---

type BrowserView struct {
	browser sandbox.Browser
}

func NewBrowserView(sb sandbox.Sandbox) *BrowserView {
	return &BrowserView{browser: sb.Browser()}
}

func (t *BrowserView) Name() string { return "browser_view" }

func (t *BrowserView) Description() string {
	return tools.NormalizeDescription(
		"View the current content of the browser page.",
		"When to use: checking latest state of previously opened pages, monitoring progress, saving screenshots.",
		"Best practices: page content is automatically provided after navigation. Use this for rechecking state.",
	)
}

func (t *BrowserView) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief": tools.StringProperty("One brief sentence to explain this action"),
	}, "brief")
}

func (t *BrowserView) Execute(ctx context.Context, args map[string]any) (string, error) {
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_view",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- browser_click ---

type BrowserClick struct {
	browser sandbox.Browser
}

func NewBrowserClick(sb sandbox.Sandbox) *BrowserClick {
	return &BrowserClick{browser: sb.Browser()}
}

func (t *BrowserClick) Name() string { return "browser_click" }

func (t *BrowserClick) Description() string {
	return tools.NormalizeDescription(
		"Click an element on the browser page.",
		"When to use: clicking page elements, triggering interactions, submitting forms.",
		"Best practices: provide either element index or coordinates. Prefer element index.",
	)
}

func (t *BrowserClick) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":           tools.StringProperty("One brief sentence to explain this action"),
		"index":           tools.IntegerProperty("Index number of the element to click"),
		"viewport_width":  map[string]any{"type": "number", "description": "Viewport width for coordinate-based clicking"},
		"viewport_height": map[string]any{"type": "number", "description": "Viewport height for coordinate-based clicking"},
		"coordinate_x":    map[string]any{"type": "number", "description": "Horizontal coordinate of click position"},
		"coordinate_y":    map[string]any{"type": "number", "description": "Vertical coordinate of click position"},
	}, "brief")
}

func (t *BrowserClick) Execute(ctx context.Context, args map[string]any) (string, error) {
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_click",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- browser_input ---

type BrowserInput struct {
	browser sandbox.Browser
}

func NewBrowserInput(sb sandbox.Sandbox) *BrowserInput {
	return &BrowserInput{browser: sb.Browser()}
}

func (t *BrowserInput) Name() string { return "browser_input" }

func (t *BrowserInput) Description() string {
	return tools.NormalizeDescription(
		"Overwrite text in an editable field on the browser page.",
		"When to use: filling content in input fields, updating form fields.",
		"Best practices: this tool clears existing text first, then inputs new text. Prefer element index over coordinates.",
	)
}

func (t *BrowserInput) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief":           tools.StringProperty("One brief sentence to explain this action"),
		"index":           tools.IntegerProperty("Index number of the element to input text into"),
		"viewport_width":  map[string]any{"type": "number", "description": "Viewport width for coordinate-based targeting"},
		"viewport_height": map[string]any{"type": "number", "description": "Viewport height for coordinate-based targeting"},
		"coordinate_x":    map[string]any{"type": "number", "description": "Horizontal coordinate of the element"},
		"coordinate_y":    map[string]any{"type": "number", "description": "Vertical coordinate of the element"},
		"text":            tools.StringProperty("Full text content to input. This will overwrite any existing content."),
		"press_enter": map[string]any{
			"type":        "boolean",
			"description": "Whether to simulate pressing Enter after input",
		},
	}, "brief", "text", "press_enter")
}

func (t *BrowserInput) Execute(ctx context.Context, args map[string]any) (string, error) {
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_input",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- browser_scroll_up ---

type BrowserScrollUp struct {
	browser sandbox.Browser
}

func NewBrowserScrollUp(sb sandbox.Sandbox) *BrowserScrollUp {
	return &BrowserScrollUp{browser: sb.Browser()}
}

func (t *BrowserScrollUp) Name() string { return "browser_scroll_up" }

func (t *BrowserScrollUp) Description() string {
	return tools.NormalizeDescription(
		"Scroll up the browser page.",
		"When to use: viewing content above, returning to page top.",
		"Best practices: use to_top parameter to scroll directly to top.",
	)
}

func (t *BrowserScrollUp) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief": tools.StringProperty("One brief sentence to explain this action"),
		"to_top": map[string]any{
			"type":        "boolean",
			"description": "If true, scrolls directly to the top of the page. Defaults to false.",
		},
	}, "brief")
}

func (t *BrowserScrollUp) Execute(ctx context.Context, args map[string]any) (string, error) {
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_scroll_up",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- browser_scroll_down ---

type BrowserScrollDown struct {
	browser sandbox.Browser
}

func NewBrowserScrollDown(sb sandbox.Sandbox) *BrowserScrollDown {
	return &BrowserScrollDown{browser: sb.Browser()}
}

func (t *BrowserScrollDown) Name() string { return "browser_scroll_down" }

func (t *BrowserScrollDown) Description() string {
	return tools.NormalizeDescription(
		"Scroll down the browser page.",
		"When to use: viewing content below, jumping to page bottom, triggering lazy-loaded content.",
		"Best practices: use to_bottom parameter to scroll directly to bottom. Multiple scrolls may be needed.",
	)
}

func (t *BrowserScrollDown) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief": tools.StringProperty("One brief sentence to explain this action"),
		"to_bottom": map[string]any{
			"type":        "boolean",
			"description": "If true, scrolls directly to the bottom of the page. Defaults to false.",
		},
	}, "brief")
}

func (t *BrowserScrollDown) Execute(ctx context.Context, args map[string]any) (string, error) {
	res, err := t.browser.Execute(ctx, sandbox.BrowserAction{
		ToolName: "browser_scroll_down",
		Params:   args,
	})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s", res.Error)
	}
	return formatBrowserResult(res), nil
}

// --- helper ---

func formatBrowserResult(res *sandbox.BrowserResult) string {
	if res.Content != "" {
		return res.Content
	}
	out := ""
	if res.Title != "" {
		out += fmt.Sprintf("Title: %s\n", res.Title)
	}
	if res.URL != "" {
		out += fmt.Sprintf("URL: %s\n", res.URL)
	}
	if out == "" {
		return "Browser action completed."
	}
	return out
}
