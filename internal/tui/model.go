package internal

import (
	"github.com/Ikit24/gomini/internal/database"
	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
)

type Model struct {
	Sessions        []database.Session
	Messages        []database.Message
	SelectedSession uuid.UUID
	MessageInput    textinput.Model
}
