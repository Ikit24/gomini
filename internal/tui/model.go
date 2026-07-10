package tui

import (
	"fmt"
	"context"
	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/google/uuid"
)

type Model struct {
	currentUser     uuid.UUID
	sessions        []database.Session
	messages        []database.Message
	selectedSession uuid.UUID
	messageInput    textinput.Model
	db              *database.DB
	geminiClient    *gemini.Client
	cancel          context.CancelFunc
	currentStream   string
	channel         chan tea.Msg
	terminalWidth   int
	viewport        viewport.Model
	errorMessage    string
	currentState    appState
	pastSessions    []database.Session
	browseCursor    int
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
		messageInput:    ti,
		db:              db,
		geminiClient:    client,
		channel:         ch,
		currentUser:     userID,
		selectedSession: uuid.Nil,
		currentState:    StateWelcome,
		pastSessions:    sessions,
		spinner:         s,
		renderer:        createMarkdownRenderer(80),
	}
}

func createMarkdownRenderer(width int) *glamour.TermRenderer {
	customStyle := styles.DarkStyleConfig
	customStyle.Document.Margin = uintPtr(0)
    customStyle.Paragraph.Margin = uintPtr(0)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStyles(customStyle),
	)
	if err != nil {
		fmt.Println("Failed to start renderer:", err)
	}
	return renderer
}

func uintPtr(i uint) *uint {
	return &i
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}
