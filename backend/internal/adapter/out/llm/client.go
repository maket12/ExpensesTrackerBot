package llm

import (
	"context"
	"fmt"

	"backend/internal/adapter/out/llm/errs"
	"backend/internal/adapter/out/llm/mapper"
	"backend/internal/domain/port"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
	tools  []openai.Tool
}

func NewClient(baseURL, apiKey, model string) (*Client, error) {
	if apiKey == "" {
		return nil, errs.ErrEmptyAPIKey
	}
	if model == "" {
		model = openai.GPT4oMini
	}

	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	return &Client{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
		tools:  mapper.DefaultOpenAITools(),
	}, nil
}

func (c *Client) GenerateResponse(ctx context.Context, messages []port.AgentMessage) (port.AgentResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    mapper.ToOpenAIMessages(messages),
		Temperature: 0.1,
		Tools:       c.tools,
		ToolChoice:  "auto",
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return port.AgentResponse{}, fmt.Errorf("%w: %v", errs.ErrCompletionFailed, err)
	}

	if len(resp.Choices) == 0 {
		return port.AgentResponse{}, errs.ErrEmptyResponse
	}

	choice := resp.Choices[0]
	return mapper.ToDomainAgentResponse(choice), nil
}
