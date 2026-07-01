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
	"github.com/muesli/reflow/wordwrap"
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
		m.TerminalWidth = msg.Width
		m.Viewport.Height = msg.Height - 3
		m.Viewport.Width = msg.Width

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
	switch m.CurrentState {
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
	var cmd, inputCmd, viewportCmd tea.Cmd
	contentChanged := false

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		return m, spinCmd

	case ArrivingMsg:
		m.isLoading = false
		m.CurrentStream += string(msg)
		cmd = waitForChunk(m.Channel)
		contentChanged = true

	case StreamFinish:
		m.isLoading = false
		finishedStream := database.Message{
			ID:        uuid.New(),
			UserID:    m.CurrentUser,
			SessionID: m.SelectedSession,
			Role:      database.ModelRole,
			Content:   m.CurrentStream,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.Messages = append(m.Messages, finishedStream)
		m.CurrentStream = ""
		contentChanged = true

		aiSaveCmd := saveMessageToDB(m.DB, finishedStream)
		cmd = tea.Batch(cmd, aiSaveCmd)

	case dbSaveErrorMsg:
		m.isLoading = false
		m.ErrorMessage = msg.err.Error()

	case dbSaveSuccessMsg:

	case geminiStreamErrorMsg:
		m.isLoading = false
		m.ErrorMessage = msg.err.Error()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			userInput := m.MessageInput.Value()
			if m.SelectedSession == uuid.Nil {
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
					UserID:    m.CurrentUser,
					Title:     title,
					CreatedAt: now,
					UpdatedAt: now,
				}
				err := m.DB.CreateSession(&newSession)
				if err != nil {
					m.ErrorMessage = "Failed to create session: %v" + err.Error()
					return m, nil
				}
				m.SelectedSession = newSessionID
				m.PastSessions = append([]database.Session{newSession}, m.PastSessions...)
			}
			geminiHistory := make([]gemini.Message, len(m.Messages))
			for i, msg := range m.Messages {
				geminiHistory[i] = gemini.Message{
					Role:    string(msg.Role),
					Content: msg.Content,
				}
			}
			dbMessage := database.Message{
				ID:        uuid.New(),
				SessionID: m.SelectedSession,
				UserID:    m.CurrentUser,
				Role:      database.UserRole,
				Content:   userInput,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			m.isLoading = true
			m.Messages = append(m.Messages, dbMessage)
			m.MessageInput.SetValue("")
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)

			cmd = waitForChunk(m.Channel)
			dbSave := saveMessageToDB(m.DB, dbMessage)
			var geminiCmd tea.Cmd
			m, geminiCmd = m.startGeminiStream(m.Channel, userInput, m.GeminiClient, geminiHistory)
			cmd = tea.Batch(cmd, dbSave, geminiCmd, inputCmd, m.spinner.Tick)
			contentChanged = true

		case "up", "down", "pgup", "pgdn":
			m.Viewport, viewportCmd = m.Viewport.Update(msg)

		case "ctrl+n":
			return m.startNewChat()

		case "ctrl+b":
			return m.switchToBrowse()

		case "ctrl+d":
			return m.deleteSelectedSession()

		default:
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)
		}
	default:
		m.Viewport, viewportCmd = m.Viewport.Update(msg)
		m.MessageInput, inputCmd = m.MessageInput.Update(msg)
	}
	if contentChanged {
		m = m.refreshViewportContent()
	}
	return m, tea.Batch(inputCmd, viewportCmd, cmd)
}

func (m Model) refreshViewportContent() Model {
	var s string
	safeWidth := m.TerminalWidth - 2
	for _, msg := range m.Messages {
		if msg.Role == database.UserRole {
			coloredPrefix := formatText(userPrefixColor, "You: ")
			s += wordwrap.String(coloredPrefix+msg.Content, safeWidth) + "\n\n"
		}
		if msg.Role == database.ModelRole {
			coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
			s += wordwrap.String(coloredPrefix+msg.Content, safeWidth) + "\n\n"
		}
	}
	if m.CurrentStream != "" {
		coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
		s += wordwrap.String(coloredPrefix+m.CurrentStream, safeWidth) + "\n"
	}
	m.Viewport.SetContent(s)
	m.Viewport.GotoBottom()
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
	m.CurrentState = StateChat
	m.SelectedSession = uuid.Nil
	m.Messages = []database.Message{}
	m.Viewport.SetContent("")
	m.MessageInput.Focus()
	m.MessageInput.Reset()
	return m, textinput.Blink
}

func (m Model) switchToBrowse() (tea.Model, tea.Cmd) {
	sessions, err := m.DB.GetSessionsByUserID(m.CurrentUser)
	if err != nil {
		m.ErrorMessage = "Failed to fetch session: " + err.Error()
		return m, nil
	}
	m.PastSessions = sessions
	m.BrowseCursor = 0
	m.CurrentState = StateBrowse
	return m, nil
}

func (m Model) deleteSelectedSession() (tea.Model, tea.Cmd) {
	var sessionToDelete uuid.UUID
	if m.CurrentState == StateBrowse {
		if len(m.PastSessions) == 0 {
			return m, nil
		}
		sessionToDelete = m.PastSessions[m.BrowseCursor].ID
	} else {
		if m.SelectedSession == uuid.Nil {
			return m, nil
		}
		sessionToDelete = m.SelectedSession
	}
	err := m.DB.DeleteSessionBySessionID(sessionToDelete)
	if err != nil {
		m.ErrorMessage = "Session deletion failed: " + err.Error()
		return m, nil
	}
	if m.CurrentState == StateBrowse {
		m.PastSessions = append(m.PastSessions[:m.BrowseCursor], m.PastSessions[m.BrowseCursor+1:]...)
		if m.BrowseCursor >= len(m.PastSessions) && m.BrowseCursor > 0 {
			m.BrowseCursor--
		}
	}
	if sessionToDelete == m.SelectedSession {
		m.SelectedSession = uuid.Nil
		m.Messages = nil
		//return user to menu
		m.CurrentState = StateBrowse
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.BrowseCursor > 0 {
				m.BrowseCursor--
			}
		case "down":
			if m.BrowseCursor < len(m.PastSessions)-1 {
				m.BrowseCursor++
			}
		case "ctrl+d":
			return m.deleteSelectedSession()
		case "esc":
			m.CurrentState = StateWelcome
			return m, nil
		case "enter":
			if len(m.PastSessions) == 0 {
			return m, nil
			}
			selectedSession := m.PastSessions[m.BrowseCursor]
			m.SelectedSession = selectedSession.ID
			messagesFromSession, err := m.DB.GetMessagesBySessionID(selectedSession.ID)
			if err != nil {
				m.ErrorMessage = "Failed to fetch messages: " + err.Error()
				return m, nil
			}
			m.Messages = messagesFromSession
			var s string
			for _, msg := range m.Messages {
				if msg.Role == database.UserRole {
					coloredPrefix := formatText(userPrefixColor, "You: ")
					s += wordwrap.String(coloredPrefix + msg.Content, m.TerminalWidth) + "\n\n"
				}
				if msg.Role == database.ModelRole {
					coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
					s += wordwrap.String(coloredPrefix + msg.Content, m.TerminalWidth) + "\n\n"
				}
			}
			if m.CurrentStream != "" {
				coloredPrefix := formatText(gominiPrefixColor, "Gemini: ")
				s += wordwrap.String(coloredPrefix + m.CurrentStream, m.TerminalWidth) + "\n"
			}
			m.Viewport.SetContent(s)
			m.Viewport.GotoBottom()
			m.CurrentState = StateChat
			m.MessageInput.Focus()
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
