package gemini

import (
	"errors"
	"strings"
	"time"
	"context"
	"google.golang.org/genai"
)

type Client struct {
	genaiClient   *genai.Client
	genaiSysTools *genai.GenerateContentConfig
	models        []string
	modelIndex    int
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

func (c *Client) CycleModel() {
	c.modelIndex++
	if c.modelIndex == len(c.models) {
		c.modelIndex = 0
	}
}

func (c *Client) CurrentModel() string {
	return c.models[c.modelIndex]
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
			genaiClient:   c,
			models:        []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.5-flash-lite"},
			modelIndex:    0,
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
	var apiErr *genai.APIError
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
	
	iter := c.genaiClient.Models.GenerateContentStream(ctx, c.CurrentModel(), sdkHistory, c.genaiSysTools)
	ch := make(chan string)
	go func() {
		defer close(ch)
		for resp, err := range iter {
			if err != nil {
				ch <- err.Error()
				//fallback model logic
				if errors.As(err, &apiErr) {
					if apiErr.Code == 429 || apiErr.Code == 503 {
						c.modelIndex = 0
						newIter := c.genaiClient.Models.GenerateContentStream(ctx, c.CurrentModel(), sdkHistory, c.genaiSysTools)
						for newResp, err := range newIter {
							if err != nil {
								ch <- err.Error()
								return
							}
							if len(newResp.Candidates) > 0 && newResp.Candidates[0].Content != nil {
								for _, part := range newResp.Candidates[0].Content.Parts {
									if part.Text != "" {
										ch <- part.Text
									}
								}
							} else if len(newResp.Candidates) > 0 {
								ch <- fmt.Sprintf("%v", newResp.Candidates[0].FinishReason)
							}
						}
					}
				}
				return
			}
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						ch <- part.Text
					} else if len(resp.Candidates) > 0 && resp.Candidates[0].Content == nil{
						ch <- fmt.Sprintf("%v", resp.Candidates[0].FinishReason)
					}

				}
			}
		}
	}()

	return ch, nil
}
