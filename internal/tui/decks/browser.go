package decks

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/tui/shared"
)

type loadedMsg struct {
	stats []*domain.DeckStats
	err   error
}

type Model struct {
	deckSvc *service.DeckService
	stats   []*domain.DeckStats
	cursor  int
	err     error
}

func New(svc *service.DeckService) Model {
	return Model{deckSvc: svc}
}

func (m Model) Init() tea.Cmd {
	return m.loadDecks()
}

func (m Model) loadDecks() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.deckSvc.AllDeckStats(context.Background())
		return loadedMsg{stats: stats, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.stats = msg.stats
		m.err = msg.err
		return m, nil
	case tea.KeyMsg:
		switch {
		case shared.Matches(msg, shared.DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Down):
			if m.cursor < len(m.stats)-1 {
				m.cursor++
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Select):
			if len(m.stats) > 0 {
				deckName := m.stats[m.cursor].DeckName
				return m, func() tea.Msg {
					return shared.SwitchScreenMsg{Screen: shared.ScreenReview, DeckName: deckName}
				}
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Browse):
			if len(m.stats) > 0 {
				deckName := m.stats[m.cursor].DeckName
				return m, func() tea.Msg {
					return shared.SwitchScreenMsg{Screen: shared.ScreenCardBrowser, DeckName: deckName}
				}
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Back):
			return m, func() tea.Msg {
				return shared.SwitchScreenMsg{Screen: shared.ScreenMenu}
			}
		case shared.Matches(msg, shared.DefaultKeyMap.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := shared.StyleTitle.Render("Decks") + "\n\n"

	if m.err != nil {
		return s + shared.StyleAgain.Render("Error: "+m.err.Error())
	}

	if len(m.stats) == 0 {
		s += shared.StyleSubtle.Render("No decks yet. Use `recurva decks new <name>` to create one.")
		return s
	}

	for i, st := range m.stats {
		cursor := "  "
		nameStyle := shared.StyleSubtle
		if i == m.cursor {
			cursor = "▶ "
			nameStyle = shared.StyleSelected
		}
		line := fmt.Sprintf("%s%s  %s",
			cursor,
			nameStyle.Render(st.DeckName),
			shared.StyleSubtle.Render(fmt.Sprintf("due: %d / total: %d", st.DueCards, st.TotalCards)),
		)
		s += line + "\n"
	}

	s += "\n" + shared.StyleHelp.Render("↑/↓ navigate • enter review • b browse • esc back • q quit")
	return s
}
