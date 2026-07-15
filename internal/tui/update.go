package tui

import (
	"strings"
	"context"
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"time"
)

var (
	userPrefixColor = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	gominiPrefixColor = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
)

type GeminiResponseMsg string
type geminiStreamErrorMsg struct {
	err error
}
type ArrivingMsg string
type StreamFinish struct{}
type ChunkChan chan tea.Msg
type dbSaveSuccessMsg struct{}
type dbSaveErrorMsg struct {
	err error
}

func waitForChunk(ch ChunkChan) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func saveMessageToDB(db *database.DB, msg database.Message) tea.Cmd {
	return func() tea.Msg {
		err := db.SaveMessage(&msg)
		if err != nil {
			return dbSaveErrorMsg{err: err}
		}
		return dbSaveSuccessMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.viewport.Height = msg.Height - 5
		m.viewport.Width = msg.Width
		//terminal message spacing
		gutterWidth := 10
		messageWidth := m.viewport.Width - gutterWidth
		if messageWidth < 10 {
			messageWidth = 10
		}
		glamourWrapWidth := messageWidth - 4
		m.renderer = createMarkdownRenderer(glamourWrapWidth)
		m = m.refreshViewportContent()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+n":
			if m.cancel != nil {
				m.cancel()
				m.cancel = nil
			}
		}
	}
	//local msg routing based on state
	switch m.currentState {
	case StateWelcome:
		return m.updateWelcome(msg)
	case StateChat:
		return m.updateChat(msg)
	case StateBrowse:
		return m.updateBrowse(msg)
	default:
		return m, nil
	}
}

func (m Model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, inputCmd, viewportCmd, geminiCmd tea.Cmd
	
	contentChanged := false

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		return m, spinCmd

	case ArrivingMsg:
		m.currentStream += string(msg)
		cmd = waitForChunk(m.channel)
		contentChanged = true

	case StreamFinish:
		m.isLoading = false
		finishedStream := database.Message{
			ID:        uuid.New(),
			UserID:    m.currentUser,
			SessionID: m.selectedSession,
			Role:      database.ModelRole,
			Content:   m.currentStream,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.messages = append(m.messages, finishedStream)
		m.currentStream = ""
		contentChanged = true

		aiSaveCmd := saveMessageToDB(m.db, finishedStream)
		cmd = tea.Batch(cmd, aiSaveCmd)

	case dbSaveErrorMsg:
		m.isLoading = false
		m.errorMessage = msg.err.Error()

	case dbSaveSuccessMsg:

	case geminiStreamErrorMsg:
		m.isLoading = false
		m.errorMessage = msg.err.Error()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.isLoading {
				return m, nil
			}

			cleanInput := strings.TrimSpace(m.messageInput.Value())
			if cleanInput == "" {
				m.messageInput.Reset()
				return m, nil
			}
			userInput := cleanInput
			m.messageInput.Reset()

			if m.selectedSession == uuid.Nil {
				title := userInput
				//dynamic chat title
				if len(title) > 35 {
					lastSpace := strings.LastIndex(title[:35], " ")
					if lastSpace != -1 {
						title = title[:lastSpace] + "..."
					} else {
						title = title[:32] + "..."
					}
				} else if title == "" {
					title = "New Chat"
				}
				newSessionID := uuid.New()
				now := time.Now().UTC()
				newSession := database.Session{
					ID:        newSessionID,
					UserID:    m.currentUser,
					Title:     title,
					CreatedAt: now,
					UpdatedAt: now,
				}
				err := m.db.CreateSession(&newSession)
				if err != nil {
					m.errorMessage = "Failed to create session: %v" + err.Error()
					return m, nil
				}
				m.selectedSession = newSessionID
				m.pastSessions = append([]database.Session{newSession}, m.pastSessions...)
			}
			geminiHistory := make([]gemini.Message, len(m.messages))
			for i, msg := range m.messages {
				geminiHistory[i] = gemini.Message{
					Role:    string(msg.Role),
					Content: msg.Content,
				}
			}
			dbMessage := database.Message{
				ID:        uuid.New(),
				SessionID: m.selectedSession,
				UserID:    m.currentUser,
				Role:      database.UserRole,
				Content:   userInput,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			m.isLoading = true
			m.messages = append(m.messages, dbMessage)
			m.messageInput.SetValue("")
			m.messageInput, inputCmd = m.messageInput.Update(msg)

			cmd = waitForChunk(m.channel)
			dbSave := saveMessageToDB(m.db, dbMessage)
			m, geminiCmd = m.startGeminiStream(m.channel, userInput, m.geminiClient, geminiHistory)
			cmd = tea.Batch(cmd, dbSave, geminiCmd, inputCmd, m.spinner.Tick)
			contentChanged = true

		case "up", "down", "pgup", "pgdn":
			m.viewport, viewportCmd = m.viewport.Update(msg)

		case "ctrl+n":
			return m.startNewChat()

		case "ctrl+b":
			return m.switchToBrowse()

		case "ctrl+d":
			return m.deleteSelectedSession()

		default:
			m.messageInput, inputCmd = m.messageInput.Update(msg)
		}
	default:
		m.viewport, viewportCmd = m.viewport.Update(msg)
		m.messageInput, inputCmd = m.messageInput.Update(msg)
	}
	if contentChanged {
		m = m.refreshViewportContent()
	}

	return m, tea.Batch(inputCmd, viewportCmd, cmd)
}

func (m Model) refreshViewportContent() Model {
	var s string
	gutterWidth := 10
	gutterStyle := lipgloss.NewStyle().Width(gutterWidth)
	messageWidth := m.viewport.Width - gutterWidth
	messageStyle := lipgloss.NewStyle().Width(messageWidth)
	
	for _, msg := range m.messages {
		if msg.Role == database.UserRole {
			renderedMessage, err := m.renderer.Render(msg.Content)
			renderedMessage = strings.TrimSpace(renderedMessage)
			if err != nil {
				m.errorMessage = "Failed to render user message: " + err.Error()
				return m
			}
			coloredPrefix := formatText(userPrefixColor, "You: ")
			boxedPrefix := gutterStyle.Render(coloredPrefix)
		    boxedMessage := messageStyle.Render(renderedMessage)
			s += lipgloss.JoinHorizontal(lipgloss.Top, boxedPrefix, boxedMessage) + "\n\n"
		}
		if msg.Role == database.ModelRole {
			renderedGominiMessage, err := m.renderer.Render(msg.Content)
			renderedGominiMessage = strings.TrimSpace(renderedGominiMessage)
			if err != nil {
				m.errorMessage = "Failed to render gemini message: " + err.Error()
				return m
			}
			coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
			boxedAiPrefix := gutterStyle.Render(coloredPrefix)
			boxedAiMessage := messageStyle.Render(renderedGominiMessage)
			s += lipgloss.JoinHorizontal(lipgloss.Top, boxedAiPrefix, boxedAiMessage) + "\n\n"
		}
	}
	if m.currentStream != "" {
		renderedGominiMessage, err := m.renderer.Render(m.currentStream)
		renderedGominiMessage = strings.TrimSpace(renderedGominiMessage)
		if err != nil {
			m.errorMessage = "Failed to stream message: " + err.Error()
			return m
		}
		coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
		boxedAiPrefix := gutterStyle.Render(coloredPrefix)
		boxedAiMessage := messageStyle.Render(renderedGominiMessage)
		s += lipgloss.JoinHorizontal(lipgloss.Top, boxedAiPrefix, boxedAiMessage) + "\n\n"
	}
	m.viewport.SetContent(s)
	m.viewport.GotoBottom()
	return m
}

func (m Model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n":
			return m.startNewChat()

		case "ctrl+b":
			return m.switchToBrowse()
		}
	}
	return m, nil
}

func (m Model) startNewChat() (tea.Model, tea.Cmd) {
	m.currentState = StateChat
	m.selectedSession = uuid.Nil
	m.messages = []database.Message{}
	m.viewport.SetContent("")
	m.messageInput.Focus()
	m.messageInput.Reset()
	return m, textinput.Blink
}

func (m Model) switchToBrowse() (tea.Model, tea.Cmd) {
	sessions, err := m.db.GetSessionsByUserID(m.currentUser)
	if err != nil {
		m.errorMessage = "Failed to fetch session: " + err.Error()
		return m, nil
	}
	m.pastSessions = sessions
	m.browseCursor = 0
	m.currentState = StateBrowse
	return m, nil
}

func (m Model) deleteSelectedSession() (tea.Model, tea.Cmd) {
	var sessionToDelete uuid.UUID
	if m.currentState == StateBrowse {
		if len(m.pastSessions) == 0 {
			return m, nil
		}
		sessionToDelete = m.pastSessions[m.browseCursor].ID
	} else {
		if m.selectedSession == uuid.Nil {
			return m, nil
		}
		sessionToDelete = m.selectedSession
	}
	err := m.db.DeleteSessionBySessionID(sessionToDelete)
	if err != nil {
		m.errorMessage = "Session deletion failed: " + err.Error()
		return m, nil
	}
	if m.currentState == StateBrowse {
		m.pastSessions = append(m.pastSessions[:m.browseCursor], m.pastSessions[m.browseCursor+1:]...)
		if m.browseCursor >= len(m.pastSessions) && m.browseCursor > 0 {
			m.browseCursor--
		}
	}
	if sessionToDelete == m.selectedSession {
		m.selectedSession = uuid.Nil
		m.messages = nil
		//return user to menu
		m.currentState = StateBrowse
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+g":
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.messageInput.Blur()
			} else {
				m.messageInput.Focus()
			}
			return m, nil

		case "up":
			if m.browseCursor > 0 {
				m.browseCursor--
			}

		case "down":
			if m.browseCursor < len(m.pastSessions)-1 {
				m.browseCursor++
			}
		
		case "ctrl+d":
			return m.deleteSelectedSession()
		
		case "esc":
			if m.showHelp {
				m.showHelp = false
				m.messageInput.Focus()
        	    return m, nil
        	}
			m.currentState = StateWelcome
			m.messageInput.Blur()
			return m, nil

		case "enter":
			if len(m.pastSessions) == 0 {
			return m, nil
			}
			selectedSession := m.pastSessions[m.browseCursor]
			m.selectedSession = selectedSession.ID
			messagesFromSession, err := m.db.GetMessagesBySessionID(selectedSession.ID)
			if err != nil {
				m.errorMessage = "Failed to fetch messages: " + err.Error()
				return m, nil
			}
			m.messages = messagesFromSession
			m = m.refreshViewportContent()
			m.currentState = StateChat
			m.messageInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m Model) startGeminiStream(ch chan tea.Msg, prompt string, client *gemini.Client, history []gemini.Message) (Model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	return m, func() tea.Msg {
		streamChan, err := client.GenerateChatResponse(ctx, history, prompt)
		if err != nil {
			return geminiStreamErrorMsg{err: err}
		}
		for text := range streamChan {
			ch <- ArrivingMsg(text)
		}
		ch <- StreamFinish{}
		return nil
	}
}
