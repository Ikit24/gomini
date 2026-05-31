package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/database"
	"google.golang.org/api/iterator"
	"github.com/google/generative-ai-go/genai"
)

type GeminiResponseMsg string
type ArrivingMsg string
type StreamFinish struct{}
type ChunkChan chan tea.Msg

func waitForChunk(ch ChunkChan) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			userInput := m.MessageInput.Value()

			dbMessage := database.Message{
				SessionID: m.SelectedSession,
				Role:      database.UserRole,
				Content:   userInput,
			}
			m.Messages = append(m.Messages, dbMessage)
			m.MessageInput.SetValue("")

			go func(ch chan tea.Msg, prompt string, client *gemini.Client) {
				stream := client.GenerateContentStream(context.Background(), prompt)
				for {
					resp, err := stream.Next()
					if err == iterator.Done {
						//server has nothing to send
						break
					}
					if err != nil {
						//exit upon network error
						break
					}

					for _, part := range resp.Candidates[0].Content.Parts {
						if text, ok := part.(genai.Text); ok {
							ch<-ArrivingMsg(string(text))
						}
					}
				}
				ch<-StreamFinish{}
			}(m.Channel, userInput, m.GeminiClient)

			return m, waitForChunk(m.Channel)
		}

		case ArrivingMsg:
			m.CurrentStream += string(msg)
			return m, waitForChunk(m.Channel)

		case StreamFinish:
			finishedStream := database.Message{
				SessionID: m.SelectedSession,
				Role:      database.ModelRole,
				Content:   m.CurrentStream,
			}
			m.Messages = append(m.Messages, finishedStream)
			m.CurrentStream = ""
			return m, nil

	case GeminiResponseMsg:
		aiMessage := database.Message{
			SessionID: m.SelectedSession,
			Role:      database.ModelRole,
			Content:   string(msg),
		}
		m.Messages = append(m.Messages, aiMessage)

		return m, nil
	}

	var cmd tea.Cmd
	m.MessageInput, cmd = m.MessageInput.Update(msg)

	return m, cmd
}

func sendToGemini(client *gemini.Client, prompt string) tea.Cmd {
	return func() tea.Msg {
		response, err := client.GenerateContent(context.Background(), prompt)
		if err != nil {
			return GeminiResponseMsg("error generating response from Gemini: " + err.Error())
		}
		return GeminiResponseMsg(response)
	}
}
