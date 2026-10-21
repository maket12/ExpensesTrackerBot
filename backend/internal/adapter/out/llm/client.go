package llm

import (
	pkgerrs "backend/pkg/errs"
	"context"
	"fmt"

	"backend/internal/domain/port"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
	tools  []openai.Tool
}

func NewClient(baseURL, apiKey, model string) (*Client, error) {
	if baseURL == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_base_url")
	}
	if apiKey == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_api_key")
	}
	if model == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_model")
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	return &Client{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
		tools:  DefaultOpenAITools(),
	}, nil
}

func (c *Client) GenerateResponse(ctx context.Context, messages []port.AgentMessage) (port.AgentResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    ToOpenAIMessages(messages),
		Temperature: 0.1,
		Tools:       c.tools,
		ToolChoice:  "auto",
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return port.AgentResponse{}, fmt.Errorf("failed to fetch llm model: %w", err)
	}

	if len(resp.Choices) == 0 {
		return port.AgentResponse{}, fmt.Errorf("expected llm to response but got nothing")
	}

	choice := resp.Choices[0]
	return ToDomainAgentResponse(choice), nil
}
