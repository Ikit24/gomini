package openai

import (
	"fmt"
	"time"
	"os"
	"errors"
	"strings"
	"context"
	"sync"
	"github.com/Ikit24/gomini/internal/llm"
	"github.com/openai/openai-go/option"
	extopenai "github.com/openai/openai-go/v3"
)

type Client struct {
	mu             sync.RWMutex
	apiKey         string
	currentModel   string
	models         []string
	preferredModel string
	persona        llm.Persona
}

func NewClient(apiKey string) *Client {
	//client := extopenai.NewClient()

	//response, err := client.Responses.New(context.Background(), responses.ResponseNewParams{
	//	Model: "gpt-4o",
	//	Input: responses.ResponseNewParamsInputUnion{
	//		OfString: extopenai.String("You are a helpful and thorough assistant in a terminal UI.")
	//	},
	//})
	//if err != nil {
	//	return err
	//}

	return &Client{
		apiKey: apiKey,
		preferredModel: "gpt-4o-mini",
		models: []string{"gpt-4o", "gpt-4o-mini"},
	}
}

func (c *Client) CurrentModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferredModel
}

func (c *Client) CycleModel() {
	c.mu.RLock()
	current := c.preferredModel
	c.mu.RUnlock()

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

func (c *Client) setModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.preferredModel = model
}

func (c *Client) attemptOrder() [] string {
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

func (c *Client) GenerateChatResponse(ctx context.Context, history []llm.Message, newPrompt string) (<-chan string, error) {
	messages := make([]extopenai.ChatCompletionMessageParamUnion, 0, len(history)+2)

	if c.persona.Prompt != "" {
		messages = append(messages, extopenai.SystemMessage(c.persona.Prompt))
	}

	for _, msg := range history {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}

		if msg.Role == "user" {
			messages = append(messages, extopenai.UserMessage(msg.Content))
		} else {
			messages = append(messages, extopenai.AssistantMessage(msg.Content))
		}
	}

	messages = append(messages, extopenai.UserMessage(newPrompt))

	params := extopenai.ChatCompletionNewParams{
		Messages:     messages,
		Temperature: extopenai.Float(float64(c.persona.Temperature)),
		TopP:        extopenai.Float(float64(c.persona.TopP)),
	}

	ch := make(chan string)

	go func() {
		var streamErr error
		var apiErr *extopenai.Error
		order := c.attemptOrder()

		for attempt := 0; attempt < len(order); attempt++ {
			model := order[attempt]
			streamErr = nil

			params.Model = extopenai.String(model)
			
			client := extopenai.NewClient(option.WithAPIKey(c.apiKey))
			stream := client.Chat.Completions.NewStreaming(ctx, params)

			for stream.Next() {
				chunk := stream.Current()
				if len(chunk.Choices) > 0 {
					delta := chunk.Choices[0].Delta.Content
					if delta != "" {
						ch <-delta
					}
				}
			}

			streamErr = stream.Err()

			if streamErr != nil {
				file, fileErr := os.OpenFile("debug.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
				if fileErr == nil {
					file.WriteString(fmt.Sprintf("Model: %s | Error Type: %T | Error: %v\n", model, streamErr, streamErr))
					file.Close()
				}
			}

			//if no err whole response stream successfull
			if streamErr == nil {
				return
			}

			if errors.As(streamErr, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode == 503 || apiErr.StatusCode == 404) {
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

func (c *Client) SetPersona(persona llm.Persona) {
	c.persona = persona
}
