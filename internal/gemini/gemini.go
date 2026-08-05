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
	genaiSysTools *genai.GenerateContentConfig
	model         string
	basePrompt    string
	personaPrompt string
	filePrompt    string
}

type Message struct {
	Role    string
	Content string
}

func(c *Client) rebuildSystemInstruction() {
	c.genaiSysTools.SystemInstruction.Parts = []*genai.Part{}
	if c.basePrompt != "" {
		c.genaiSysTools.SystemInstruction.Parts = append(
			c.genaiSysTools.SystemInstruction.Parts,
			&genai.Part{Text: c.basePrompt})
	}

	if c.personaPrompt != "" {
		c.genaiSysTools.SystemInstruction.Parts = append(
			c.genaiSysTools.SystemInstruction.Parts,
			&genai.Part{Text: c.personaPrompt})
	}

	if c.filePrompt != "" {
		c.genaiSysTools.SystemInstruction.Parts = append(
			c.genaiSysTools.SystemInstruction.Parts,
			&genai.Part{Text: c.filePrompt})
	}
}

func (c *Client) SetPersona(personaText string) {
	c.personaPrompt = personaText
	c.rebuildSystemInstruction()
}

func NewClient(ctx context.Context, apiKey string, fileContent string) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	//date&time hallucination
	currentDate := time.Now().Format("01-02-2006")

	client := &Client{
			genaiClient: c,
			model:       "gemini-2.5-flash",
			genaiSysTools: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{},
				Tools: []*genai.Tool{
					{GoogleSearch: &genai.GoogleSearch{}},
				},
			},
			basePrompt:  "You are a helpful and thorough assistant in a terminal UI. The current date is: " + currentDate,
			filePrompt:  fileContent,
		}

	client.rebuildSystemInstruction()

	return client, nil
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
