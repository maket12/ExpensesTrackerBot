package dto

type ToolCallDTO struct {
	ID        string
	Name      string
	Arguments string
}

type AgentMessageDTO struct {
	Role       string
	Content    string
	ToolCalls  []ToolCallDTO
	ToolCallID string
}

type AgentToolDTO struct {
	Name        string
	Description string
	Parameters  any
}

type GenerateCompletionRequest struct {
	Model    string
	Messages []AgentMessageDTO
	Tools    []AgentToolDTO
}

type GenerateCompletionResponse struct {
	Content   string
	ToolCalls []ToolCallDTO
}
