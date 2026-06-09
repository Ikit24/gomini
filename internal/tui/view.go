package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
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
