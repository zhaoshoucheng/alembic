package tools

import (
	"context"
	"strings"

	llmstypes "github.com/shoucheng/my-first-agent/internal/llm/langchaingo/llms"
)

type JSONSchema map[string]any

type Tool interface {
	Name() string
	Description() string
	Parameters() JSONSchema
	Execute(ctx context.Context, args map[string]any) (string, error)
}

func Definition(tool Tool) llmstypes.Tool {
	return llmstypes.Tool{
		Type: "function",
		Function: &llmstypes.FunctionDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		},
	}
}

func ObjectSchema(properties map[string]any, required ...string) JSONSchema {
	schema := JSONSchema{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		req := make([]any, 0, len(required))
		for _, name := range required {
			req = append(req, name)
		}
		schema["required"] = req
	}
	return schema
}

func StringProperty(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func IntegerProperty(description string) map[string]any {
	return map[string]any{"type": "integer", "description": description}
}

func NormalizeDescription(parts ...string) string {
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
