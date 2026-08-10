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

	sessListPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#874BFD"))
)

func formatText(style lipgloss.Style, text string) string {
	return style.Render(text)
}

func (m Model) View() string {
	if m.errorMessage != "" {
		return formatText(tooltipPrefix, "critical error: ") + m.errorMessage + formatText(tooltipPrefix, "\nPress [ctrl+c] to quit.")
	}

	if m.showHelp {
		return m.helpView()
	}

	var currentView string
	switch m.currentState {
	case StateWelcome:
		currentView = m.viewWelcome()
	case StateChat:
		currentView = m.viewChat()
	case StateBrowse:
		currentView = m.viewBrowse()
	default:
		currentView = "Unknown application state"
	}

	return currentView
}

func (m Model) helpView() string {
	var helpBoxStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#18ffa2")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#18ffa2")).
		Margin(1, 2).
		Width(m.terminalWidth - 4).
		Height(m.terminalHeight - 4).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	var helpInfoStyle = lipgloss.NewStyle().
		Width(35).
		Align(lipgloss.Left)

	help := "[ctrl+g] or [esc] Close this menu.\n\n" +
		"[ctrl+b] Browse your history.\n\n" +
		"[ctrl+n] Start new chat.\n\n" +
		"[ctrl+y] Copy the latest code block.\n\n" +
		"[alt+y] Copy the latest LLM message.\n\n" +
		"[alt+0] Switch to general-purpose persona, concise, readable answers.\n\n" +
		"[alt+1] Switch to Socratic coding tutor.\n\n" +
		"[alt+2] Switch to value investor by the rules of B. Graham and W. Buffett.\n\n" +
		"[ctrl+c] Quit application.\n\n" +
		"[ctrl+d] Delete selected sessions Warning!!! This is instant and cannot be reversed.\n\n" +
		"You can use navigation when in a session,for ex. using [ctrl+b] will return you to the session list.\n\n"

	help = helpInfoStyle.Render(help)

	return helpBoxStyle.Render(help)
}

func (m Model) viewBrowse() string {
	var chatsHeader, savedChats string

	var chatQuitStyle = lipgloss.NewStyle().
		Width(45).
		Align(lipgloss.Center).Bold(true)

	var chatInfoStyle = lipgloss.NewStyle().
		Width(70).
		Align(lipgloss.Center)

	var chatsBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Margin(1, 2).
		Width(m.terminalWidth - 4).
		Height(m.terminalHeight - 4).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	chatsHeader += formatText(tooltipPrefix, "Previous chats:") + "\n"

	for i, session := range m.pastSessions {
		if i == m.browseCursor {
			savedChats += selectedStyle.Render(fmt.Sprintf("-> [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		} else {
			savedChats += unselectedStyle.Render(fmt.Sprintf("   [CreatedAt: %s] Title: %s", session.CreatedAt.Format("02/01/2006"), session.Title)) + "\n"
		}
	}

	savedChats += formatText(chatQuitStyle,"\nPress [esc] to return")
	savedChats = chatInfoStyle.Render(savedChats)

	return chatsBoxStyle.Render(chatsHeader + "\n" + savedChats)
}

func (m Model) viewWelcome() string {
	var title, s, sessList string

	var welcomeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Margin(1, 2).
		Width(m.terminalWidth - 4).
		Height(m.terminalHeight - 4).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	title += formatText(tooltipPrefix, "Welcome to Gomini!\n")

	s += formatText(tooltipPrefix,
		"Press [ctrl+g] for help.\n\n" +
		"Press [ctrl+n] to start new chat.\n\n" +
		"Press [ctrl+b] to browse your history.\n\n")

	if len(m.pastSessions) > 0 {
		sessList += formatText(sessListPrefix,"You have " + fmt.Sprint(len(m.pastSessions)) + " previous conversations.")
	}
	s += formatText(tooltipPrefix, "\nPress [ctrl+c] to quit.")

	return welcomeBoxStyle.Render(title + "\n\n" + sessList + "\n\n" + s)
}

func (m Model) viewChat() string {
	var UI string = m.viewport.View() + "\n"

	if m.isLoading {
		UI += m.spinner.View() + " Looking for answers...\n\n"
	}

	if m.errorMessage != "" {
		var errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)
	UI += errorStyle.Render(m.errorMessage) + "\n"
	}

	var chatInputStyle = lipgloss.NewStyle().
        Width(m.terminalWidth).
        Align(lipgloss.Center)
    UI += chatInputStyle.Render(m.messageInput.View()) + "\n"

	halfWidth := m.terminalWidth / 2

    personaStyle := lipgloss.NewStyle().
        Width(halfWidth - 10).
        Align(lipgloss.Center).
        Foreground(lipgloss.Color("#874BFD"))

    statusStyle := lipgloss.NewStyle().
        Width(halfWidth).
        Align(lipgloss.Left).
        Foreground(lipgloss.Color("#18ffa2"))

    statusMsgText := statusStyle.Render(m.statusMessage)
	combinedLeft := fmt.Sprintf("%s | %s", m.activePersona, m.geminiClient.CurrentModel())
	combinedLeftText := personaStyle.Render(combinedLeft)

    statusBar := lipgloss.JoinHorizontal(lipgloss.Top, combinedLeftText, statusMsgText)

    UI += statusBar

    return UI
}
