package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

type Model struct {
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
}

func InitialModel(db *database.DB, client *gemini.Client) Model {
	ch := make(chan tea.Msg)
	ti := textinput.New()
	ti.Placeholder = "Please enter your message..."
	ti.Focus()

	sessionID := uuid.New()
	
	sess := &database.Session{
		ID:     sessionID,
		UserID: uuid.Nil,
		Title:  "Local Chat",
	}
	err := db.SaveSession(sess)
	if err != nil {
		panic("SaveSession failed: " + err.Error())
	}

	return Model{
		MessageInput:    ti,
		DB:              db,
		GeminiClient:    client,
		Channel:         ch,
		SelectedSession: sessionID,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
