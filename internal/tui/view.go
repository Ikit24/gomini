package tui

import (
	"github.com/Ikit24/gomini/internal/database"
)

func (m Model) View() string {
	var s string
	for _, msg := range m.Messages {
		if msg.Role == database.UserRole {
			s += "You: " + msg.Content + "\n"
		}
		if msg.Role == database.ModelRole {
			s += "Gemini: " + msg.Content + "\n"
		}
	}
	inputBox := m.MessageInput.View()

	return s + "\n" + inputBox
}
