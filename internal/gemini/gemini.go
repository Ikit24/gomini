package gemini

import {
	"context"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
}

type Client struct {
	genaiClient *genai.Client
	model       *genai.GenerativeModel
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {

}
