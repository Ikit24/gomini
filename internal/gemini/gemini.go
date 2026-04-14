package gemini

import (
	"context"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type Client struct {
	genaiClient *genai.Client
	model       *genai.GenerativeModel
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
