package tui

func (m Model) View() string {
	inputBox := m.MessageInput.View()

	return m.LastMessage + "\n" + inputBox
}
