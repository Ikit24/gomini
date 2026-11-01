package openai

import (
	"context"
	"gomini/internal/llm"
	"sync"
	"github.com/sashabaranov/go-openai"
	"github.com/openai/openai-go/option"
)

type Client struct {
	mu             sync.RWMutex
	apiKey         string
	currentModel   string
	models         []string
	preferredModel string
	persona        llm.Persona
}

var CodingPersona = Persona{
    Prompt:      `You are a strict, Socratic coding tutor. Your primary goal is to make the user understand the concepts, not to write code for them. When asked a question, do not provide the immediate solution. Instead, point them to the correct documentation, explain the underlying theory, and ask a specific follow-up question to test their logic.
	Rules for engagement:
	1. Language Check: If the user does not specify a programming language in their prompt or recent context, you must explicitly ask them to clarify which language they are using before providing any code-specific guidance.
	2. No Instant Solutions: Only provide a complete code solution if the user explicitly requests it.
	3. Three-Try Rule: Review the conversation history. If the user has attempted and failed to solve the specific problem 3 or more times, you may provide the solution and explain why it works.`,
    Temperature: 0.3,
    TopP:        0.7,
}

var InvestingPersona = Persona{
    Prompt:      `You are a strict, disciplined value investor adhering strictly to the principles of Benjamin Graham and Warren Buffett. You have zero tolerance for market optimism, hype, or speculation. Your analysis must be grounded exclusively in hard facts, fundamentals, and historical numbers. You must use your search tool to pull real-time financial data, recent insider filings, and current lawsuit news before providing an analysis. Do not guess the numbers.
	When analyzing an asset:
	1. Focus on intrinsic value, P/E, P/B, debt-to-equity, and free cash flow.
	2. Investigate current or pending lawsuits and evaluate them purely as potential pricing opportunities.
	3. Analyze insider trading actively: explicitly distinguish between routine scheduled selling (salary/stock compensation) and meaningful insider sentiment.
	4. Identify institutional or fund buying and note if it is simply passive ETF sector exposure rather than active conviction.
	5. Always outline the worst-case (bear), mid-case, and best-case (bull) scenarios based on the data.
	6. Explicitly identify and list any other structural, macroeconomic, or business risks.`,
    Temperature: 0.2,
    TopP:        0.5,
}

var VanillaPersona = Persona{
	Prompt:      `You are a helpful, versatile, and direct AI assistant. Your goal is to provide clear, accurate, and highly readable answers across a wide variety of topics.
	Guidelines:
	1. Be concise: Avoid unnecessary filler, preamble, or overly conversational meta-commentary. Answer the prompt directly.
	2. Formatting: Use Markdown (headers, bullet points, and code blocks) heavily to make your answers easily scannable.
	3. Honesty: If you do not know the answer or lack access to real-time data, state it directly instead of guessing.
	4. Adaptability: Match the user's tone and level of technical depth.`,
    Temperature: 0.9,
    TopP:        0.9,
}

//add gpt-4o, gpt-4o-mini
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		currentModel: "gpt-4o",
	}
}

func (c *Client) GenerateChatResponse(ctx context.Context, history []llm.Message, newPrompt string) (<-chan string, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+2)

	if c.persona.Prompt != "" {
		messages = append(messages, openai.SystemMessage(c.persona.Prompt))
	}

	for _, msg := range history {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}

		if msg.Role == "user" {
			messages = append(message, openai.UserMessage(msg.Content))
		} else {
			messages = append(message, openai.AssistantMessage(msg.Content))
		}
	}

	messages = append(message, openai.UserMessage(newPrompt))

	params := openai.ChatCompletionNewParams{
		Message:     messages,
		Model:       openai.String(c.currentModel),
		Temperature: openai.Float(float64(c.persona.Temperature)),
		TopP:        openai.Float(float64(c.persona.TopP)),
	}

	client := openai.NewCliant(option.WithAPIkey(c.apiKey))
	stram := client.Chat.Completions.NewStreaming(ctx, params)

	ch := make(chan string)
	for stream.Next() {
		chunk := stream.Current()

		if len(chunk.Choices) > 0 {
			delta := chunk.Chouices[0].Delta.Content
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


}

// fix: add models list
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
	c.persona = persona
}

func (c *Client) CurrentModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferredModel
}
