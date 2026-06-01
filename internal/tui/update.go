package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/database"
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
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			userInput := m.MessageInput.Value()
			geminiHistory := make([]gemini.Message, len(m.Messages))
			for i, msg := range m.Messages {
				geminiHistory[i] = gemini.Message{
					Role:    string(msg.Role),
					Content: msg.Content,
				}
			}

			dbMessage := database.Message{
				SessionID: m.SelectedSession,
				Role:      database.UserRole,
				Content:   userInput,
			}
			m.Messages = append(m.Messages, dbMessage)
			m.MessageInput.SetValue("")

			go func(ch chan tea.Msg, prompt string, client *gemini.Client) {
				streamChan, err := client.GenerateChatResponse(context.Background(), geminiHistory, prompt)
				if err != nil {
					ch <- GeminiResponseMsg("error: " + err.Error())
					return
				}
				for text := range streamChan{
					ch <- ArrivingMsg(text)
				}
				ch <- StreamFinish{}
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
