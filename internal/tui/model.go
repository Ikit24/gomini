package tui

import (
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type Model struct {
	CurrentUser     uuid.UUID
	Sessions        []database.Session
	Messages        []database.Message
	SelectedSession uuid.UUID
	MessageInput    textinput.Model
	LastMessage     string
	DB              *database.DB
	GeminiClient    *gemini.Client
	CurrentStream   string
	Channel         chan tea.Msg
	TerminalWidth   int
	Viewport        viewport.Model
	ErrorMessage    string
	CurrentState    appState
	PastSessions    []database.Session
	BrowseCursor    int
}

type appState int

const (
	StateWelcome appState = iota
	StateChat
	StateBrowse
)

func InitialModel(db *database.DB, client *gemini.Client, userID uuid.UUID, sessions []database.Session) Model {
	ch := make(chan tea.Msg)
	ti := textinput.New()
	ti.Placeholder = "Please enter your message..."
	ti.Focus()

	return Model{
		MessageInput:    ti,
		DB:              db,
		GeminiClient:    client,
		Channel:         ch,
		CurrentUser:     userID,
		SelectedSession: uuid.Nil,
		CurrentState:    StateWelcome,
		PastSessions:    sessions,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
