package cards

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/tui/shared"
)

type savedMsg struct{ err error }

type field int

const (
	fieldFront field = iota
	fieldBack
	fieldNotes
)

type Model struct {
	cardSvc  *service.CardService
	deckName string
	front    textinput.Model
	back     textarea.Model
	notes    textinput.Model
	active   field
	err      error
	saved    bool
}

func New(svc *service.CardService, deckName string) Model {
	front := textinput.New()
	front.Placeholder = "Front (question)"
	front.Focus()

	back := textarea.New()
	back.Placeholder = "Back (answer)"

	notes := textinput.New()
	notes.Placeholder = "Notes (optional)"

	return Model{
		cardSvc:  svc,
		deckName: deckName,
		front:    front,
		back:     back,
		notes:    notes,
	}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case savedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.saved = true
			m.front.SetValue("")
			m.back.SetValue("")
			m.notes.SetValue("")
			m.active = fieldFront
			m.front.Focus()
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.active = (m.active + 1) % 3
			m.front.Blur()
			m.back.Blur()
			m.notes.Blur()
			switch m.active {
			case fieldFront:
				m.front.Focus()
			case fieldBack:
				m.back.Focus()
			case fieldNotes:
				m.notes.Focus()
			}
			return m, nil
		case "ctrl+s":
			return m, m.save()
		case "esc":
			return m, func() tea.Msg { return shared.SwitchScreenMsg{Screen: shared.ScreenMenu} }
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.front, cmd = m.front.Update(msg)
	cmds = append(cmds, cmd)
	m.back, cmd = m.back.Update(msg)
	cmds = append(cmds, cmd)
	m.notes, cmd = m.notes.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) save() tea.Cmd {
	front := strings.TrimSpace(m.front.Value())
	back := strings.TrimSpace(m.back.Value())
	notes := strings.TrimSpace(m.notes.Value())
	return func() tea.Msg {
		_, err := m.cardSvc.AddCard(context.Background(), m.deckName, front, back, notes, nil)
		return savedMsg{err: err}
	}
}

func (m Model) View() string {
	s := shared.StyleTitle.Render("Add Card") + " "
	s += shared.StyleSubtle.Render("Deck: "+m.deckName) + "\n\n"

	s += "Front:\n" + m.front.View() + "\n\n"
	s += "Back:\n" + m.back.View() + "\n\n"
	s += "Notes:\n" + m.notes.View() + "\n\n"

	if m.saved {
		s += shared.StyleGood.Render("✓ Card saved!") + "\n"
	}
	if m.err != nil {
		s += shared.StyleAgain.Render("Error: "+m.err.Error()) + "\n"
	}

	s += shared.StyleHelp.Render("tab to switch fields • ctrl+s to save • esc back")
	return s
}
