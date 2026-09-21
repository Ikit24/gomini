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
	preferredModel  string
	basePrompt      string
	personaPrompt   string
	filePrompt      string
}

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

var CodingPersona = Persona{
    Prompt:      `You are a strict, Socratic coding tutor. Your primary goal is to make the user understand the concepts, not to write code for them. When asked a question, do not provide the immediate solution. Instead, point them to the correct documentation, explain the underlying theory, and ask a specific follow-up question to test their logic.
	Rules for engagement:
	1. Language Check: If the user does not specify a programming language in their prompt or recent context, you must explicitly ask them to clarify which language they are using before providing any code-specific guidance.
	2. No Instant Solutions: Only provide a complete code solution if the user explicitly requests it.
	3. Three-Try Rule: Review the conversation history. If the user has attempted and failed to solve the specific problem 3 or more times, you may provide the solution and explain why it works.`,
    Temperature: 0.3,
    TopP:        0.7,
    TopK:        40,
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
    TopK:        40,
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
    TopK:        50,
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

func (c *Client) SetPersona(persona Persona) {
	c.personaPrompt = persona.Prompt
	c.genaiSysTools.Temperature = genai.Ptr[float32](persona.Temperature)
	c.genaiSysTools.TopP = genai.Ptr[float32](persona.TopP)
	c.genaiSysTools.TopK = genai.Ptr[float32](persona.TopK)
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

	// Uncomment lines below to see what your API key can see
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
			genaiSysTools:  &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{},
				Temperature: genai.Ptr[float32](0.7),
				TopP:        genai.Ptr[float32](0.95),
				TopK:        genai.Ptr[float32](40),
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
					file, fileErr := os.OpenFile("debug.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
					//check on this ferr
					if fileErr == nil {
						file.WriteString(fmt.Sprintf("Model: %s | Error Type: %T | Error: %v\n", model, streamErr, streamErr))
						file.Close()
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
