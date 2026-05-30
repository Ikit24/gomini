package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Ikit24/gomini/internal/gemini"
)

type GeminiResponseMsg string

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
			return m, sendToGemini(m.GeminiClient, userInput)
		}
	case GeminiResponseMsg:
		m.LastMessage = string(msg)

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
