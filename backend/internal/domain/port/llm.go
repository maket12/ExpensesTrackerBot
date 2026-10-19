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

type ToolParameterProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolParameters struct {
	Type       string                            `json:"type"`
	Properties map[string]ToolParameterProperty `json:"properties"`
	Required   []string                          `json:"required,omitempty"`
}

type AgentTool struct {
	Name        string
	Description string
	Parameters  ToolParameters
}

type AgentResponse struct {
	Content   string
	ToolCalls []ToolCall
}

type LLMClient interface {
	GenerateResponse(ctx context.Context, messages []AgentMessage, tools []AgentTool) (AgentResponse, error)
}
