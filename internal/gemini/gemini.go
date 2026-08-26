package gemini

import (
	"os"
	"fmt"
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
			models:        []string{"gemini-2.5-flash", "gemini-3.5-flash-lite", "gemini-3.1-pro-preview"},
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

func processResponse (ch chan string, resp *genai.GenerateContentResponse) {
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				ch <- part.Text
			}
		}
	} else if len(resp.Candidates) > 0 && resp.Candidates[0].Content == nil{
		ch <- fmt.Sprintf("%v", resp.Candidates[0].FinishReason)
	}
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
	
	ch := make(chan string)
	go func() {
		defer close(ch)
		var streamErr error
		for attempt := 0; attempt < len(c.models); attempt++ {
			iter := c.genaiClient.Models.GenerateContentStream(ctx, c.CurrentModel(), sdkHistory, c.genaiSysTools)
	
			for resp, err := range iter {
				if err != nil {
					f, _ := os.OpenFile("debug.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
					f.WriteString(err.Error() + "\n")
					f.Close()
					streamErr = err
					break
				}
				processResponse(ch, resp)
			}
			if streamErr == nil {
				//whole response streamed successfully
				return
			}
			//fallback model logic
			if errors.As(streamErr, &apiErr) && (apiErr.Code == 429 || apiErr.Code == 503 || apiErr.Code == 404) {
				c.CycleModel()
			} else {
				ch <- streamErr.Error()
				return
			}
		}
		ch <- streamErr.Error()
	}()

	return ch, nil
}
