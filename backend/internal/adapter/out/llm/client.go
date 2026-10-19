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
	}, nil
}

func (c *Client) GenerateResponse(ctx context.Context, messages []port.AgentMessage, tools []port.AgentTool) (port.AgentResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    mapper.ToOpenAIMessages(messages),
		Temperature: 0.1,
	}

	if len(tools) > 0 {
		req.Tools = mapper.ToOpenAITools(tools)
		req.ToolChoice = "auto"
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return port.AgentResponse{}, fmt.Errorf("%w: %v", errs.ErrCompletionFailed, err)
	}

	if len(resp.Choices) == 0 {
		return port.AgentResponse{}, errs.ErrEmptyResponse
	}

	return mapper.ToDomainAgentResponse(resp.Choices[0]), nil
}
