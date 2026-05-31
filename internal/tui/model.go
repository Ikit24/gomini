package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
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
}

func InitialModel(db *database.DB, client *gemini.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()

	return Model{
		MessageInput: ti,
		DB:           db,
		GeminiClient: client,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
