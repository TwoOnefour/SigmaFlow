package gemini

import (
	"context"
	"github.com/twoonefour/sigmaflow/pkg/llm"
	"google.golang.org/genai"
)

type Client struct {
	client         *genai.Client
	model          string
	thinkingBudget *int32
}

func NewClient(apiKey string, model string, thinkingBudget int32) (*Client, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, err
	}

	return &Client{client, model, &thinkingBudget}, nil
}

func (c *Client) Chat(ctx context.Context, messages []llm.Messages) (string, error) {
	msg := make([]*genai.Content, 0)
	var systemPrompt string
	for _, m := range messages {
		switch m.Role {
		case llm.RoleUser:
			msg = append(msg, &genai.Content{
				Parts: []*genai.Part{
					{Text: m.Content},
				},
				Role: genai.RoleUser,
			})
		case llm.RoleSystem:
			systemPrompt = m.Content
		case llm.RoleAssistant:
			msg = append(msg, &genai.Content{
				Parts: []*genai.Part{
					{Text: m.Content},
				},
				Role: genai.RoleModel,
			})
		}
	}
	config := &genai.GenerateContentConfig{}
	config.ThinkingConfig = &genai.ThinkingConfig{
		ThinkingBudget: c.thinkingBudget,
	}
	if systemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(systemPrompt, genai.RoleUser)
	}
	res, err := c.client.Models.GenerateContent(ctx, c.model, msg, config)
	if err != nil {
		return "", err
	}
	return res.Text(), nil
}
