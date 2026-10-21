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

type GenerateCompletionRequest struct {
	Model    string
	Messages []AgentMessageDTO
	Tools    []AgentToolDTO
}

type GenerateCompletionResponse struct {
	Content   string
	ToolCalls []ToolCallDTO
}
