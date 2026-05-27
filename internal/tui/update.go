package tui

import (
   tea "github.com/charmbracelet/bubbletea"
)

type GeminiResponseMsg string

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			userInput := m.MessageInput.Value()
			m.MessageInput.Value("")
			return m, nil
		}
	case GeminiResponseMsg:
		m.LastMessage = string(msg)

		return m, nil
	}

	var cmd tea.Cmd
	m.MessageInput, cmd = m.MessageInput.Update(msg)

	return m, cmd
}
