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
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#18ffa2")).Bold(true)
)

func formatText(style lipgloss.Style, text string) string {
	return style.Render(text)
}

func (m Model) View() string {
	if m.errorMessage != "" {
		return formatText(tooltipPrefix, "Critical error: ") + m.errorMessage + formatText(tooltipPrefix, "\nPress [ctrl+c] to quit.")
	}

	if m.showHelp {
		return m.helpView()
	}

	switch m.currentState {
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

func (m Model) helpView() string {
	return formatText(helpStyle, "[ctrl+g] Close this menu.\n" +
		"[ctrl+b] Browse your history.\n" +
		"[ctrl+n] Start new chat.\n" +
		"[ctrl+c] Quit application.\n" +
		"[ctrl+d] Delete selected sessions (Warning!!! This is instant and cannot be reversed).\n" +
		"You can use navigation when in a session,for ex. using [ctrl+b] will return you to the session list.\n")
}

func (m Model) viewBrowse() string {
	var savedChats string
	savedChats += formatText(tooltipPrefix, "Saved Chats:") + "\n\n"

	for i, session := range m.pastSessions {
		if i == m.browseCursor {
			savedChats += selectedStyle.Render(fmt.Sprintf("-> [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		} else {
			savedChats += unselectedStyle.Render(fmt.Sprintf("   [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		}
	}
	savedChats += formatText(tooltipPrefix, "\nPress [esc] to return")
	return savedChats
}

func (m Model) viewWelcome() string {
	var s string
	var welcomeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Margin(1, 2).
		Width(m.terminalWidth - 4).
		Height(m.terminalHeight - 4).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	s += formatText(tooltipPrefix, "Welcome to Gomini!\n\n")

	if len(m.pastSessions) > 0 {
		s += "You have " + fmt.Sprint(len(m.pastSessions)) + " previous conversations.\n\n"
		s += formatText(tooltipPrefix,
			"Press [ctrl+g] for help.\n" +
			"Press [ctrl+n] to start new chat.\n" +
			"Press [ctrl+b] to browse your history.\n")
	} else {
		s += formatText(tooltipPrefix, "Press [ctrl+n] to start new chat.")
	}
	s += formatText(tooltipPrefix, "\n\nPress [ctrl+c] to quit.")

	return welcomeBoxStyle.Render(s)
}

func (m Model) viewChat() string {
	var errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	var UI string = m.viewport.View() + "\n"
	if m.isLoading {
		UI += m.spinner.View() + " Looking for answers...\n\n"
	}

	if m.errorMessage != "" {
		UI += errorStyle.Render(m.errorMessage) + "\n"
	}
	UI += m.messageInput.View()

	return UI
}
