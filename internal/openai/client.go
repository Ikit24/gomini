package openai

import (
	"context"
	"gomini/internal/llm"
)

type Client struct {
	apiKey       string
	currentModel string
	person       llm.Persona
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		currentModel: "gpt-4o",
	}
}

func (c *Client) GenerateChatResponse(ctx context.Context, history []llm.Message, newPrompt string) (<-chan string, error) {
	stream := make(chan string)

	//implement openai stream
	close(stream)

	return stream, nil
}

//add CycleModel
//add SetPersona
//add CurrentModle
