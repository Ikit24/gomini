package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#88C0D0")).
		Foreground(lipgloss.Color("#1e1e2e")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241"))
	tooltipPrefix = lipgloss.NewStyle().Bold(true)
)

func formatText(style lipgloss.Style, text string) string {
	return style.Render(text)
}

func (m Model) View() string {
	if m.ErrorMessage != "" {
		return formatText(tooltipPrefix, "Critical error: ") + m.ErrorMessage + formatText(tooltipPrefix, "\nPress [ctrl+c] to quit.")
	}
	switch m.CurrentState {
	case StateWelcome:
		return m.viewWelcome()
	case StateChat:
		return m.viewChat()
	case StateBrowse:
		return m.viewBrowse()
	default:
		return "Unknown application state"
	}
}

func (m Model) viewBrowse() string {
	var savedChats string
	savedChats += formatText(tooltipPrefix, "Saved Chats:") + "\n\n"

	for i, session := range m.PastSessions {
		if i == m.BrowseCursor {
			savedChats += selectedStyle.Render(fmt.Sprintf("->    [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		} else {
			savedChats += unselectedStyle.Render(fmt.Sprintf("   [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		}
	}
	savedChats += formatText(tooltipPrefix, "\nPress [esc] to return")
	return savedChats
}

func (m Model) viewWelcome() string {
	var s string
	s += formatText(tooltipPrefix, "Welcome to Gomini! \n\n")

	if len(m.PastSessions) > 0 {
		s += "You have " + fmt.Sprint(len(m.PastSessions)) + " previous conversations.\n"
		s += formatText(tooltipPrefix, "Press [ctrl+b] to browse your history, or [ctrl+n] to start new chat.")
	} else {
		s += formatText(tooltipPrefix, "Press [ctrl+n] to start new chat.")
	}
	s += formatText(tooltipPrefix, "\n\nPress [ctrl+c] to quit.")
	//lip gloss later here to center text and add borders
	return s
}

func (m Model) viewChat() string {
	var errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	var UI string = m.Viewport.View() + "\n"
	if m.isLoading {
		UI += m.spinner.View() + " Contemplating life choices...\n\n"
	}

	if m.ErrorMessage != "" {
		UI += errorStyle.Render(m.ErrorMessage) + "\n"
	}
	UI += m.MessageInput.View()

	return UI
}
