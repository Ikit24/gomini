package gemini

import (
	"strings"
	"time"
	"context"
	"fmt"
	"google.golang.org/genai"
)

type Client struct {
	genaiClient   *genai.Client
	model         string
	genaiSysTools *genai.GenerateContentConfig
}

type Message struct {
	Role    string
	Content string
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	//date&time hallucination
	currentDate := time.Now().Format("01-02-2006")
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text:"You are a helfpul and through assistant in a terminal UI. The current date is: " + currentDate},
			},
		},
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
	}

	return &Client{
		genaiClient:   c,
		model:         "gemini-2.5-flash",
		genaiSysTools: config,
	}, nil
}

func (c *Client) GenerateChatResponse(ctx context.Context, history []Message, newPrompt string) (<-chan string, error) {
	sdkHistory := make([]*genai.Content, 0, len(history)+1)
	for _, msg := range history {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		sdkMsg := &genai.Content{
			Role:  msg.Role,
			Parts: []*genai.Part{{Text: msg.Content}},
		}
		sdkHistory = append(sdkHistory, sdkMsg)
	}

	newMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: newPrompt}},
	}
	sdkHistory = append(sdkHistory, newMsg)
	
	iter := c.genaiClient.Models.GenerateContentStream(ctx, c.model, sdkHistory, c.genaiSysTools)
	ch := make(chan string)
	go func() {
		defer close(ch)
		for resp, err := range iter {
			if err != nil {
				fmt.Println("couldn't stream data", err)
				return
			}
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						ch <- part.Text
					}
				}
			}
		}
	}()
	return ch, nil
}
