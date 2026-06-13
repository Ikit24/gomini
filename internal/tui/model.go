package tui

import (
	"log"
	"time"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
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
}

type appState int

const (
	StateWelcome appState = iota
	StateChat
	StateBrowse
)

func InitialModel(db *database.DB, client *gemini.Client, userID uuid.UUID) Model {
	ch := make(chan tea.Msg)
	ti := textinput.New()
	ti.Placeholder = "Please enter your message..."
	ti.Focus()

	sessionID := uuid.New()
	
	sess := &database.Session{
		ID:        sessionID,
		UserID:    userID,
		Title:     "Local Chat",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.SaveSession(sess)
	if err != nil {
		log.Fatalf("Failed to save session: %v", err)
	}

	return Model{
		MessageInput:    ti,
		DB:              db,
		GeminiClient:    client,
		Channel:         ch,
		CurrentUser:     userID,
		SelectedSession: sessionID,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
