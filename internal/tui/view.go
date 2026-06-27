package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var cursorPrefixColor = lipgloss.NewStyle().Foreground(lipgloss.Color("123")).Bold(true)

func (m Model) View() string {
	if m.ErrorMessage != "" {
		return "Critical error: " + m.ErrorMessage + "\nPress [ctrl+c] to quit."
	}
	switch m.CurrentState {
	case StateWelcome:
		return m.viewWelcome()
	case StateChat:
		return m.viewChat()
	case StateBrowse:
		var savedChats string
		savedChats += "Saved Chats:\n\n"
		for i, session := range m.PastSessions {
			if i == m.BrowseCursor {
				coloredPrefix := cursorPrefixColor.Render("> ")
				savedChats += coloredPrefix
			} else {
				savedChats += "  "
			}
			savedChats += "SessionID: " + session.ID.String() + " " + "CreatedAt: " + session.CreatedAt.Format("02/01/2006") + "\n"
		}
		savedChats += "\nPress [esc] to return"
		return savedChats
	default:
		return "Unknown application state"
	}
}

func (m Model) viewWelcome() string {
	var s string
	s += "Welcome to Gomini! \n\n"

	if len(m.PastSessions) > 0 {
		s += "You have " + fmt.Sprint(len(m.PastSessions)) + " previous conversations.\n"
		s += "Press [ctrl+b] to browse your history, or [ctrl+n] to start new chat."
	} else {
		s += "Press [ctrl+n] to start new chat."
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
