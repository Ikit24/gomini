package tui

import (
	"time"
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/database"
)

type GeminiResponseMsg string
type geminiStreamErrorMsg struct{
	err error
}
type ArrivingMsg string
type StreamFinish struct{}
type ChunkChan chan tea.Msg
type dbSaveSuccessMsg struct{}
type dbSaveErrorMsg struct{
	err error
}

func waitForChunk(ch ChunkChan) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func saveMessageToDB(db *database.DB, msg database.Message) tea.Cmd {
	return func() tea.Msg{
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
		}
	}

	//local msg routing based on state
	switch m.CurrentState {
	case StateWelcome:
		return m.updateWelcome(msg)
	case StateChat:
		return m.updateChat(msg)
	default:
		return m, nil
	}
}

func (m Model) updateChat (msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, inputCmd, viewportCmd tea.Cmd
    contentChanged := false
	
	switch msg := msg.(type) {
	case ArrivingMsg:
		m.CurrentStream += string(msg)
		cmd = waitForChunk(m.Channel)
		contentChanged = true

	case StreamFinish:
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
			m.ErrorMessage = msg.err.Error()
		case dbSaveSuccessMsg:

		case geminiStreamErrorMsg:
			m.ErrorMessage = msg.err.Error()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			userInput := m.MessageInput.Value()
			if m.SelectedSession == uuid.Nil {
				title := userInput
				if len(title) > 35 {
					lastSpace := strings.LastIndex(title:[35], " ")
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
			m.Messages = append(m.Messages, dbMessage)
			m.MessageInput.SetValue("")
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)

			cmd = waitForChunk(m.Channel)
			dbSave := saveMessageToDB(m.DB, dbMessage)
			geminiStream  := startGeminiStream(m.Channel, userInput, m.GeminiClient, geminiHistory)
			cmd = tea.Batch(cmd, dbSave, geminiStream, inputCmd)
			contentChanged = true
			
		case "up", "down", "pgup", "pgdn":
			m.Viewport, viewportCmd = m.Viewport.Update(msg)
		
		default:
			m.MessageInput, inputCmd = m.MessageInput.Update(msg)
		}
	default:
		m.Viewport, viewportCmd = m.Viewport.Update(msg)
		m.MessageInput, inputCmd = m.MessageInput.Update(msg)
	}

	if contentChanged {
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
		m.Viewport.SetContent(s)
		m.Viewport.GotoBottom()
	}
	return m, tea.Batch(inputCmd, viewportCmd, cmd)
}

func (m Model) updateWelcome (msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			m.CurrentState = StateChat
			//old session clear to start new chat
			m.SelectedSession = uuid.Nil
			m.Messages = []database.Message{}
			m.Viewport.SetContent("")

			m.MessageInput.Focus()
			return m, textinput.Blink
			
		case "b":
			m.CurrentState = StateBrowse
			return m, nil
		}
	}
	return m, nil
}

func startGeminiStream (ch chan tea.Msg, prompt string, client *gemini.Client, history []gemini.Message) tea.Cmd {
	return func() tea.Msg {
		streamChan, err := client.GenerateChatResponse(context.Background(), history, prompt)
		if err != nil {
			return geminiStreamErrorMsg{err: err}
		}
		for text := range streamChan{
		ch <- ArrivingMsg(text)
		}
		ch <- StreamFinish{}
		return nil
	}
}
