package main

import (
	"context"
	"fmt"

	"github.com/shoucheng/my-first-agent/domain/llm"
	llmstypes "github.com/shoucheng/my-first-agent/internal/llm/langchaingo/llms"
	"github.com/shoucheng/my-first-agent/internal/tools"
	"github.com/shoucheng/my-first-agent/pkg/types"
)

const defaultSystemPrompt = "你是一个可以使用工具完成任务的智能体。需要工具时调用可用工具，得到观察结果后再给用户最终答案。"

type Agent struct {
	llm           *llm.Service
	model         string
	tools         *tools.Registry
	maxIterations int
	systemPrompt  string
	verbose       bool
}

func NewAgent(llmSvc *llm.Service, model string, registry *tools.Registry, cfg types.AgentConfig) (*Agent, error) {
	if llmSvc == nil {
		return nil, fmt.Errorf("agent.New: nil llm service")
	}
	if model == "" {
		return nil, fmt.Errorf("agent.New: model is required")
	}
	if registry == nil {
		registry = tools.NewRegistry()
	}
	maxIterations := cfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}
	return &Agent{
		llm:           llmSvc,
		model:         model,
		tools:         registry,
		maxIterations: maxIterations,
		systemPrompt:  defaultSystemPrompt,
		verbose:       cfg.Verbose,
	}, nil
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	messages := []llmstypes.MessageContent{
		llmstypes.TextParts(llmstypes.ChatMessageTypeSystem, a.systemPrompt),
		llmstypes.TextParts(llmstypes.ChatMessageTypeHuman, input),
	}

	for iteration := 0; iteration < a.maxIterations; iteration++ {
		resp, err := a.llm.GenerateContent(
			ctx,
			a.model,
			messages,
			llmstypes.WithTools(a.tools.Definitions()),
			llmstypes.WithToolChoice("auto"),
		)
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("agent.Run: empty LLM response")
		}
		choice := resp.Choices[0]
		toolCalls := normalizeToolCalls(choice)
		if len(toolCalls) == 0 {
			return choice.Content, nil
		}

		messages = append(messages, assistantToolCallMessage(choice.Content, toolCalls))
		results := a.tools.ExecuteToolCalls(ctx, toolCalls)
		for _, result := range results {
			if a.verbose {
				fmt.Printf("[tool] %s -> %s\n", result.Name, result.Content)
			}
			messages = append(messages, tools.ResultMessage(result))
		}
	}

	return "", fmt.Errorf("agent.Run: reached max iterations %d without final answer", a.maxIterations)
}

func normalizeToolCalls(choice *llmstypes.ContentChoice) []llmstypes.ToolCall {
	if len(choice.ToolCalls) > 0 {
		return choice.ToolCalls
	}
	if choice.FuncCall == nil {
		return nil
	}
	return []llmstypes.ToolCall{
		{
			ID:           "call_0",
			Type:         "function",
			FunctionCall: choice.FuncCall,
		},
	}
}

func assistantToolCallMessage(content string, calls []llmstypes.ToolCall) llmstypes.MessageContent {
	parts := make([]llmstypes.ContentPart, 0, len(calls)+1)
	if content != "" {
		parts = append(parts, llmstypes.TextPart(content))
	}
	for _, call := range calls {
		parts = append(parts, call)
	}
	return llmstypes.MessageContent{
		Role:  llmstypes.ChatMessageTypeAI,
		Parts: parts,
	}
}
