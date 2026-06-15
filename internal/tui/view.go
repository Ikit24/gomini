package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	switch m.CurrentState {
	case StateWelcome:
		return m.viewWelcome()
	case StateChat:
		return m.viewChat()
	default:
		return "Unknown application state"
	}
}

func (m Model) viewWelcome() string {
	var s string

	s += "Welcome to Gomini! \n\n"
	if len(m.PastSessions) > 0 {
		s += "You have " + fmt.Sprint(len(m.PastSessions)) + " previous conversations.\n"
		s += "Press [b] to browse your history, or [n] to start new chat."
	} else {
		s += "Press [n] to start new chat."
	}
	s += "\n\nPress [ctrl+c] to quit."

	//lip gloss later here to center text and add borders
	return s
}

func (m Model) viewChat() string {
	var errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	var UI string = m.Viewport.View() + "\n"

	if m.ErrorMessage != "" {
		UI += errorStyle.Render(m.ErrorMessage) + "\n"
	}
	UI += m.MessageInput.View()

	return UI
}
