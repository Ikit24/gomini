package tui

import (
	"github.com/muesli/reflow/wordwrap"
	"github.com/Ikit24/gomini/internal/database"
)

func (m Model) View() string {
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

	inputBox := m.MessageInput.View()
	m.Viewport.SetContent(s)

	return m.Viewport.View() + "\n" + inputBox
}
