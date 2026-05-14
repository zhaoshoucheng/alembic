package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"

	jsonrepair "github.com/RealAlexandreAI/json-repair"
	llmstypes "github.com/shoucheng/my-first-agent/internal/llm/langchaingo/llms"
)

type ParseStatus string

const (
	ParseStatusOK        ParseStatus = "ok"
	ParseStatusUnknown   ParseStatus = "unknown_tool"
	ParseStatusWrongArgs ParseStatus = "wrong_args"
)

type ParsedCall struct {
	ID        string
	Name      string
	Arguments map[string]any
	Raw       llmstypes.ToolCall
	Status    ParseStatus
	Err       error
}

type CallResult struct {
	ID      string
	Name    string
	Content string
	IsError bool
	Err     error
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("register tool: nil tool")
	}
	name := tool.Name()
	if name == "" {
		return fmt.Errorf("register tool: empty name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tools[name]; ok {
		return fmt.Errorf("register tool %q: duplicated tool name", name)
	}
	r.tools[name] = tool
	return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Definitions() []llmstypes.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	defs := make([]llmstypes.Tool, 0, len(names))
	for _, name := range names {
		defs = append(defs, Definition(r.tools[name]))
	}
	return defs
}

func (r *Registry) ParseToolCall(call llmstypes.ToolCall) ParsedCall {
	if call.FunctionCall == nil {
		return ParsedCall{
			ID:     call.ID,
			Raw:    call,
			Status: ParseStatusWrongArgs,
			Err:    fmt.Errorf("tool call has no function payload"),
		}
	}
	name := call.FunctionCall.Name
	tool, ok := r.Get(name)
	if !ok {
		return ParsedCall{
			ID:     call.ID,
			Name:   name,
			Raw:    call,
			Status: ParseStatusUnknown,
			Err:    fmt.Errorf("unknown tool %q", name),
		}
	}

	args, err := parseArguments(call.FunctionCall.Arguments)
	if err != nil {
		return ParsedCall{
			ID:     call.ID,
			Name:   name,
			Raw:    call,
			Status: ParseStatusWrongArgs,
			Err:    err,
		}
	}
	if err := validateArguments(tool.Parameters(), args); err != nil {
		return ParsedCall{
			ID:        call.ID,
			Name:      name,
			Arguments: args,
			Raw:       call,
			Status:    ParseStatusWrongArgs,
			Err:       err,
		}
	}
	return ParsedCall{
		ID:        call.ID,
		Name:      name,
		Arguments: args,
		Raw:       call,
		Status:    ParseStatusOK,
	}
}

func (r *Registry) ExecuteToolCall(ctx context.Context, call llmstypes.ToolCall) CallResult {
	parsed := r.ParseToolCall(call)
	if parsed.Status != ParseStatusOK {
		return CallResult{
			ID:      parsed.ID,
			Name:    parsed.Name,
			Content: parsed.Err.Error(),
			IsError: true,
			Err:     parsed.Err,
		}
	}
	tool, _ := r.Get(parsed.Name)
	content, err := tool.Execute(ctx, parsed.Arguments)
	if err != nil {
		return CallResult{
			ID:      parsed.ID,
			Name:    parsed.Name,
			Content: err.Error(),
			IsError: true,
			Err:     err,
		}
	}
	return CallResult{
		ID:      parsed.ID,
		Name:    parsed.Name,
		Content: content,
	}
}

func (r *Registry) ExecuteToolCalls(ctx context.Context, calls []llmstypes.ToolCall) []CallResult {
	results := make([]CallResult, 0, len(calls))
	for _, call := range calls {
		results = append(results, r.ExecuteToolCall(ctx, call))
	}
	return results
}

func ResultMessage(result CallResult) llmstypes.MessageContent {
	isError := result.IsError
	return llmstypes.MessageContent{
		Role: llmstypes.ChatMessageTypeTool,
		Parts: []llmstypes.ContentPart{
			llmstypes.ToolCallResponse{
				ToolCallID: result.ID,
				Name:       result.Name,
				Content:    result.Content,
				IsError:    &isError,
			},
		},
	}
}

func parseArguments(raw string) (map[string]any, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(raw), &args); err == nil {
		if args == nil {
			args = map[string]any{}
		}
		return args, nil
	}

	repaired, repairErr := jsonrepair.RepairJSON(raw)
	if repairErr != nil {
		return nil, fmt.Errorf("parse tool arguments: %w", repairErr)
	}
	if err := json.Unmarshal([]byte(repaired), &args); err != nil {
		return nil, fmt.Errorf("parse repaired tool arguments: %w", err)
	}
	if args == nil {
		args = map[string]any{}
	}
	return args, nil
}

func validateArguments(schema JSONSchema, args map[string]any) error {
	if len(schema) == 0 {
		return nil
	}
	if schemaType(schema) != "" && schemaType(schema) != "object" {
		return fmt.Errorf("root schema type must be object, got %q", schemaType(schema))
	}

	properties, _ := schema["properties"].(map[string]any)
	for _, name := range schemaRequired(schema) {
		if _, ok := args[name]; !ok {
			return fmt.Errorf("missing required argument %q", name)
		}
	}

	if additional, ok := schema["additionalProperties"].(bool); ok && !additional {
		for name := range args {
			if _, ok := properties[name]; !ok {
				return fmt.Errorf("unknown argument %q", name)
			}
		}
	}

	for name, value := range args {
		prop, ok := properties[name].(map[string]any)
		if !ok {
			continue
		}
		if err := validateValue(name, value, prop); err != nil {
			return err
		}
	}
	return nil
}

func schemaType(schema map[string]any) string {
	t, _ := schema["type"].(string)
	return t
}

func schemaRequired(schema map[string]any) []string {
	raw, ok := schema["required"]
	if !ok {
		return nil
	}
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if s, ok := value.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func validateValue(name string, value any, schema map[string]any) error {
	typ := schemaType(schema)
	if typ == "" {
		return nil
	}
	switch typ {
	case "string":
		if _, ok := value.(string); !ok {
			return typeError(name, typ, value)
		}
	case "number":
		if _, ok := AsFloat(value); !ok {
			return typeError(name, typ, value)
		}
	case "integer":
		number, ok := AsFloat(value)
		if !ok || math.Trunc(number) != number {
			return typeError(name, typ, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return typeError(name, typ, value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return typeError(name, typ, value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return typeError(name, typ, value)
		}
	default:
		return fmt.Errorf("argument %q uses unsupported schema type %q", name, typ)
	}
	return nil
}

func AsFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

func typeError(name, want string, got any) error {
	return fmt.Errorf("argument %q must be %s, got %T", name, want, got)
}
