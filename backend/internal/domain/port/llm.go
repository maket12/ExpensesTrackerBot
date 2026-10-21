package port

import "context"

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type AgentMessage struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

type AgentResponse struct {
	Content   string
	ToolCalls []ToolCall
}

type LLMClient interface {
	GenerateResponse(ctx context.Context, messages []AgentMessage) (AgentResponse, error)
}
