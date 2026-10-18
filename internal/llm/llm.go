package llm

import (
	"context"
)

type Message struct {
	Role    string
	Content string
}

type Persona struct {
	Prompt      string
    Temperature float32
    TopP        float32
    TopK        float32
}

type ChatProvider interface {
	GenerateChatResponse(ctx context.Context, history []Message, newPrompt string) (<-chan string, error)
}
