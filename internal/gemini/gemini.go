package gemini

import (
	"os"
	"fmt"
	"errors"
	"strings"
	"time"
	"context"
	"google.golang.org/genai"
	"sync"
)

type Client struct {
	genaiClient     *genai.Client
	mu              sync.RWMutex
	genaiSysTools   *genai.GenerateContentConfig
	models          []string
	modelIndex      int
	preferredModel  string
	fallbackChain   []string
	basePrompt      string
	personaPrompt   string
	filePrompt      string
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

func (c *Client) setModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.preferredModel = model
}

func (c *Client) CurrentModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferredModel
}

func (c *Client) CycleModel() {
	//c.mu.RLock()
	current := c.preferredModel
	//defer c.mu.RUnlock()

	currentIndex := -1
	for i, m := range c.models {
		if m == current {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(c.models)
	c.setModel(c.models[nextIndex])
}

func (c *Client) attemptOrder() []string {
	c.mu.RLock()
	preferred := c.preferredModel
	c.mu.RUnlock()

	order := []string{preferred}
	for _, m := range c.models {
		if m != preferred {
			order = append(order, m)
		}
	}
	return order
}

func NewClient(ctx context.Context, apiKey string, fileContent string) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	// Temporary check to inspect what your specific API key can see
    //resp, err := c.Models.List(ctx, nil)
    //if err != nil {
    //    return nil, fmt.Errorf("failed to list models: %w", err)
    //}
    //for _, m := range resp.Items {
    //   fmt.Printf("Authorized Model -> Name: %s\n", m.Name)
    //}

	//date&time hallucination
	currentDate := time.Now().Format("01-02-2006")

	client := &Client{
			genaiClient:    c,
			preferredModel: "gemini-2.5-flash",
			models:         []string{"gemini-3.8-flash", "gemini-3.1-flash-lite", "gemini-2.5-flash"},
			fallbackChain:  []string{"gemini-3.1-flash-lite", "gemini-3.8-flash"},
			genaiSysTools:  &genai.GenerateContentConfig{
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
	var apiErr genai.APIError

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
		order := c.attemptOrder()
		var streamErr error

		for attempt := 0; attempt < len(order); attempt++ {
			model := order[attempt]
			streamErr = nil
			iter := c.genaiClient.Models.GenerateContentStream(ctx, model, sdkHistory, c.genaiSysTools)
	
			for resp, err := range iter {
				if err != nil {
					streamErr = err
					f, ferr := os.OpenFile("debug.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
					//check on this ferr
					if ferr == nil {
						f.WriteString(fmt.Sprintf("Model: %s | Error Type: %T | Error: %v\n", model, streamErr, streamErr))
						f.Close()
					}
					break
				}
				processResponse(ch, resp)
			}

			//whole response streamed successfully
			if streamErr == nil {
				return
			}

			//fallback model logic
			if errors.As(streamErr, &apiErr) && (apiErr.Code == 429 || apiErr.Code == 503 || apiErr.Code == 404) {
				select {
				case <- ctx.Done():
					return
				case <- time.After(2 * time.Second):
					if attempt+1 < len(order) {
						fmt.Sprintf("falling back from %s to %s after error: %v", model, order[attempt+1], streamErr)
					} else {
						fmt.Sprintf("out of fallback models, giving up after error: %v", streamErr)
					}
				}
			} else {
				ch <- streamErr.Error()
				return
			}
		}
		if streamErr != nil {
			ch <- streamErr.Error()
		}
	}()

	return ch, nil
}
