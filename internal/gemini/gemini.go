package gemini

import (
	"context"
	"fmt"
	"strings"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
    "google.golang.org/api/iterator"
)

type Client struct {
	genaiClient *genai.Client
	model       *genai.GenerativeModel
}

type Message struct {
	Role    string
	Content string
}

func (c *Client) GenerateContentStream(ctx context.Context, prompt string) *genai.GenerateContentResponseIterator {
    return c.model.GenerateContentStream(ctx, genai.Text(prompt))
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

	m := c.GenerativeModel("gemini-2.5-flash")

	return &Client{
		genaiClient: c,
		model:       m,
	}, nil
}

func (c *Client) Close() error {
	return c.genaiClient.Close()
}

func (c *Client) GenerateChatResponse(ctx context.Context, history []Message, newPrompt string) (<-chan string, error) {
	cs := c.model.StartChat()

	sdkHistory := make([]*genai.Content, 0, len(history))
	for _, msg := range history {
		sdkMsg := &genai.Content{
			Role: msg.Role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		sdkHistory = append(sdkHistory, sdkMsg)
	}
	cs.History = sdkHistory

	iter := cs.SendMessageStream(ctx, genai.Text(newPrompt))
	ch := make(chan string)
	go func() {
		defer close(ch)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				//stream finished
				break
			}
			if err != nil {
				fmt.Println("couldn't stream data")
				return
			}
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil  {
				for _, part := range resp.Candidates[0].Content.Parts {
					if text, ok := part.(genai.Text); ok {
						ch <- string(text)
					}
				}
			}
		}
	}()

	return ch, nil
}
