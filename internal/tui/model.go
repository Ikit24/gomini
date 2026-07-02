package tui

import (
	"context"
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
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
	cancel          context.CancelFunc
	CurrentStream   string
	Channel         chan tea.Msg
	TerminalWidth   int
	Viewport        viewport.Model
	ErrorMessage    string
	CurrentState    appState
	PastSessions    []database.Session
	BrowseCursor    int
	spinner         spinner.Model
	isLoading       bool
	renderer        *glamour.TermRenderer
}

type appState int

const (
	StateWelcome appState = iota
	StateChat
	StateBrowse
)

func InitialModel(db *database.DB, client *gemini.Client, userID uuid.UUID, sessions []database.Session) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
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
		spinner:         s,
		renderer:        createMarkdownRenderer(TerminalWidth),
	}
}

func createMarkdownRenderer(width int) *glamour.TermRenderer {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
		)
	if err != nil {
		return fmt.Println("Failed to start renderer:", err)
	}
	return renderer
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}
