package gemini

import (
	"context"
	"google.golang.org/genai"
	"okx/internal/service/llm"
)

type Client struct {
	client *genai.Client
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (c *Client) Chat(ctx context.Context, messages []llm.Messages) (string, error) {
	thinkingBudgetVal := int32(32768)
	var systemPrompt string
	systemPrompt = messages[0].Content
	msg := messages[1].Content
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingBudget: &thinkingBudgetVal,
		},
	}
	res, err := c.client.Models.GenerateContent(ctx, "gemini-2.5-pro", genai.Text(msg), config)
	if err != nil {
		return "", err
	}
	return res.Text(), nil
}
