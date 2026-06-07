package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
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
	var cmd tea.Cmd
	var contentChanged bool
	var inputCmd, viewportCmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.Viewport.Height = msg.Height - 3
		m.Viewport.Width = msg.Width

		contentChanged = true

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
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)

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

			cmd = waitForChunk(m.Channel)
			contentChanged = true
		}

	case ArrivingMsg:
		m.CurrentStream += string(msg)
		cmd = waitForChunk(m.Channel)
		contentChanged = true

	case StreamFinish:
		finishedStream := database.Message{
			SessionID: m.SelectedSession,
			Role:      database.ModelRole,
			Content:   m.CurrentStream,
		}
		m.Messages = append(m.Messages, finishedStream)
		m.CurrentStream = ""
		contentChanged = true

	case GeminiResponseMsg:
		aiMessage := database.Message{
			SessionID: m.SelectedSession,
			Role:      database.ModelRole,
			Content:   string(msg),
		}
		m.Messages = append(m.Messages, aiMessage)
		contentChanged = true
	}

	if contentChanged {
		var s string
		for _, msg := range m.Messages {
			if msg.Role == database.UserRole {
				s += "You: " + wordwrap.String(msg.Content, m.TerminalWidth) + "\n"
			}
			if msg.Role == database.ModelRole {
				s += "Gemini: " + wordwrap.String(msg.Content, m.TerminalWidth) + "\n"
			}
		}

		if m.CurrentStream != "" {
			s += "Gemini: " + wordwrap.String(m.CurrentStream, m.TerminalWidth) + "\n"
		}

		m.Viewport.SetContent(s)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":

		case "up", "down", "pgup", "pgdn":
			m.Viewport, viewportCmd = m.Viewport.Update(msg)
		
		default:
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)
		}
	default:
		m.Viewport, viewportCmd = m.Viewport.Update(msg)
		m.MessageInput, inputCmd = m.MessageInput.Update(msg)
	}
	return m, tea.Batch(inputCmd, viewportCmd, cmd)
}
