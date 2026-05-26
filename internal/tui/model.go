package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
)

type Model struct {
	Sessions        []database.Session
	Messages        []database.Message
	SelectedSession uuid.UUID
	MessageInput    textinput.Model
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()

	return Model{
		MessageInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
