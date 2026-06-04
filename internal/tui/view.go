package tui

func (m Model) View() string {
	inputBox := m.MessageInput.View()
	return m.Viewport.View() + "\n" + inputBox
}
