package tui

func (m Model) View() string {
	header := "Hello from the TUI!"
	inputBox := m.MessageInput.View()

	return header + inputBox
}
