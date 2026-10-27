package openai

import (
	"context"
	"gomini/internal/llm"
	"sync"
)

type Client struct {
	mu             sync.RWMutex
	apiKey         string
	currentModel   string
	models         []string
	preferredModel string
	persona        llm.Persona
}

//add gpt-4o, gpt-4o-mini
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

func (c *Client) CycleModel() {
	current := c.preferredModel

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

func (c *Client) SetPersona(persona Persona) {
	c.personaPrompt = persona.Prompt
	c.genaiSysTools.Temperature = genai.Ptr[float32](persona.Temperature)
	c.genaiSysTools.TopP = genai.Ptr[float32](persona.TopP)
	c.genaiSysTools.TopK = genai.Ptr[float32](persona.TopK)
	c.rebuildSystemInstruction()
}

func (c *Client) CurrentModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferredModel
}
