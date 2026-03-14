package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r3g/recurva/internal/tui/shared"
)

type Model struct {
	items  []string
	cursor int
}

var menuItems = []string{
	"Review Cards",
	"Browse Decks",
	"Quit",
}

func New() Model {
	return Model{items: menuItems}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case shared.Matches(msg, shared.DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Select):
			return m, m.selectItem()
		case shared.Matches(msg, shared.DefaultKeyMap.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) selectItem() tea.Cmd {
	switch m.cursor {
	case 0: // Review Cards
		return func() tea.Msg {
			return shared.SwitchScreenMsg{Screen: shared.ScreenDecks}
		}
	case 1: // Browse Decks
		return func() tea.Msg {
			return shared.SwitchScreenMsg{Screen: shared.ScreenDecks}
		}
	case 2: // Quit
		return tea.Quit
	}
	return nil
}

func (m Model) View() string {
	s := shared.StyleTitle.Render("Recurva") + "\n"
	s += shared.StyleSubtle.Render("Aiming to bend the forgetting curve.") + "\n\n"

	for i, item := range m.items {
		cursor := "  "
		style := shared.StyleSubtle
		if i == m.cursor {
			cursor = "▶ "
			style = shared.StyleSelected
		}
		s += cursor + style.Render(item) + "\n"
	}

	s += "\n" + shared.StyleHelp.Render("↑/↓ navigate • enter select • q quit")
	return s
}
