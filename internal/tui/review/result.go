package review

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/tui/shared"
)

type ResultModel struct {
	summary domain.SessionSummary
}

func NewResult(summary domain.SessionSummary) ResultModel {
	return ResultModel{summary: summary}
}

func (m ResultModel) Init() tea.Cmd { return nil }

func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if shared.Matches(msg, shared.DefaultKeyMap.Quit) || shared.Matches(msg, shared.DefaultKeyMap.Select) {
			return m, func() tea.Msg { return shared.SwitchScreenMsg{Screen: shared.ScreenMenu} }
		}
	}
	return m, nil
}

func (m ResultModel) View() string {
	s := m.summary
	out := shared.StyleTitle.Render("Session Complete!") + "\n\n"
	out += fmt.Sprintf("  Total reviewed: %d\n\n", s.Total)
	out += "  " + shared.StyleAgain.Render(fmt.Sprintf("Again: %d", s.Again)) + "\n"
	out += "  " + shared.StyleHard.Render(fmt.Sprintf("Hard:  %d", s.Hard)) + "\n"
	out += "  " + shared.StyleGood.Render(fmt.Sprintf("Good:  %d", s.Good)) + "\n"
	out += "  " + shared.StyleEasy.Render(fmt.Sprintf("Easy:  %d", s.Easy)) + "\n\n"

	mins := int(s.Duration.Minutes())
	secs := int(s.Duration.Seconds()) % 60
	out += shared.StyleSubtle.Render(fmt.Sprintf("Time: %dm%ds", mins, secs)) + "\n\n"
	out += shared.StyleHelp.Render("enter/q to return to menu")
	return out
}
