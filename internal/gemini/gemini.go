package gemini

import (
	"context"
	"fmt"
	"strings"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type Client struct {
	genaiClient *genai.Client
	model       *genai.GenerativeModel
}

func(c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates returned from Gemini")
	}

	var builder strings.Builder

	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			builder.WriteString(string(text))
		}
	}

	finalResponse := builder.String()
	if finalResponse == "" {
		return "", fmt.Errorf("AI returned a response but no text content was found")
	}

	return finalResponse, nil
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	m := c.GenerativeModel("gemini-1.5-flash")

	return &Client{
		genaiClient: c,
		model:       m,
	}, nil
}

func (c *Client) Close() error {
	return c.genaiClient.Close()
}
