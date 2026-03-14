package review

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/scheduler"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/tui/shared"
)

type sessionLoadedMsg struct {
	session    *domain.ReviewSession
	priorStats ratingCounts
	err        error
}

type ratingCounts struct {
	again, hard, good, easy int
}

type ratedMsg struct {
	err error
}

type ReviewState int

const (
	ReviewStateLoading ReviewState = iota
	ReviewStateFront
	ReviewStateBack
	ReviewStateDone
)

type Model struct {
	reviewSvc  *service.ReviewService
	deckName   string
	session    *domain.ReviewSession
	state      ReviewState
	preview    *scheduler.Preview
	priorStats ratingCounts
	width      int
	err        error
}

func New(svc *service.ReviewService, deckName string) (Model, tea.Cmd) {
	m := Model{
		reviewSvc: svc,
		deckName:  deckName,
		state:     ReviewStateLoading,
	}
	return m, m.loadSession()
}

func (m Model) Init() tea.Cmd {
	return m.loadSession()
}

func (m Model) loadSession() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		session, err := m.reviewSvc.StartSession(ctx, m.deckName)
		if err != nil {
			return sessionLoadedMsg{err: err}
		}
		a, h, g, e := m.reviewSvc.PriorRatingCounts(ctx, m.deckName, 30)
		prior := ratingCounts{again: a, hard: h, good: g, easy: e}
		return sessionLoadedMsg{session: session, priorStats: prior}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case sessionLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.session = msg.session
		m.priorStats = msg.priorStats
		if m.session.Done() {
			m.state = ReviewStateDone
			return m, switchToResult()
		}
		m.state = ReviewStateFront
		return m, m.loadPreview()

	case ratedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.session.Done() {
			m.state = ReviewStateDone
			return m, switchToResult()
		}
		m.state = ReviewStateFront
		return m, m.loadPreview()

	case *scheduler.Preview:
		m.preview = msg
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case ReviewStateLoading, ReviewStateDone:
			// no key handling in these states
		case ReviewStateFront:
			switch {
			case shared.Matches(msg, shared.DefaultKeyMap.Flip):
				m.state = ReviewStateBack
			case shared.Matches(msg, shared.DefaultKeyMap.Back):
				return m, switchToMenu()
			case shared.Matches(msg, shared.DefaultKeyMap.Quit):
				return m, tea.Quit
			}
		case ReviewStateBack:
			switch {
			case shared.Matches(msg, shared.DefaultKeyMap.Again):
				return m, m.rate(domain.RatingAgain)
			case shared.Matches(msg, shared.DefaultKeyMap.Hard):
				return m, m.rate(domain.RatingHard)
			case shared.Matches(msg, shared.DefaultKeyMap.Good):
				return m, m.rate(domain.RatingGood)
			case shared.Matches(msg, shared.DefaultKeyMap.Easy):
				return m, m.rate(domain.RatingEasy)
			case shared.Matches(msg, shared.DefaultKeyMap.Back):
				m.state = ReviewStateFront
			case shared.Matches(msg, shared.DefaultKeyMap.Quit):
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func switchToResult() tea.Cmd {
	return func() tea.Msg { return shared.SwitchScreenMsg{Screen: shared.ScreenResult} }
}

func switchToMenu() tea.Cmd {
	return func() tea.Msg { return shared.SwitchScreenMsg{Screen: shared.ScreenMenu} }
}

func (m Model) loadPreview() tea.Cmd {
	card := m.session.CurrentCard()
	if card == nil {
		return nil
	}
	return func() tea.Msg {
		p, _ := m.reviewSvc.Preview(*card)
		return &p
	}
}

func (m Model) rate(rating domain.Rating) tea.Cmd {
	return func() tea.Msg {
		err := m.reviewSvc.Rate(context.Background(), m.session, rating)
		return ratedMsg{err: err}
	}
}

func (m Model) Summary() domain.SessionSummary {
	if m.session == nil {
		return domain.SessionSummary{}
	}
	return m.reviewSvc.Summary(m.session)
}

func (m Model) View() string {
	if m.err != nil {
		return shared.StyleAgain.Render("Error: " + m.err.Error())
	}

	switch m.state {
	case ReviewStateLoading:
		return shared.StyleSubtle.Render("Loading session...")

	case ReviewStateFront, ReviewStateBack:
		card := m.session.CurrentCard()
		if card == nil {
			return ""
		}
		cur, total := m.session.Progress()
		s := shared.StyleProgress.Render(fmt.Sprintf("[%d/%d]", cur, total))
		s += "  " + shared.StyleSubtle.Render("Deck: "+m.deckName) + "\n"

		// Constrain card width to terminal (border=2 + padding=4 + margin for centering)
		cardWidth := 60
		if m.width > 0 {
			cardWidth = m.width - 4
			if cardWidth > 80 {
				cardWidth = 80
			}
		}
		innerWidth := cardWidth - 6 // border (2) + padding (2*2)

		cardStyle := shared.StyleCard.Width(innerWidth)

		centeredFront := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).
			Bold(true).Foreground(shared.ColorFront).Render(card.Front)
		cardContent := centeredFront
		if m.state == ReviewStateBack {
			divider := strings.Repeat("─", innerWidth)
			cardContent += "\n\n" + shared.StyleSubtle.Render(divider) + "\n\n"
			cardContent += shared.StyleBack.Render(card.Back)
			if card.Notes != "" {
				cardContent += "\n\n" + shared.StyleSubtle.Render("Notes: "+card.Notes)
			}
		}
		s += cardStyle.Render(cardContent) + "\n"

		if m.state == ReviewStateFront {
			s += shared.StyleHelp.Render("space to flip • esc back • q quit")
		} else {
			s += renderRatingBar(m.preview)
		}

		s += "\n\n" + m.renderSessionStats()
		return s

	case ReviewStateDone:
		return shared.StyleSubtle.Render("Session complete!")
	}
	return ""
}

func renderRatingBar(preview *scheduler.Preview) string {
	s := ""
	if preview != nil {
		for _, rp := range preview.Ratings {
			interval := formatInterval(rp.Interval)
			switch rp.Rating {
			case domain.RatingAgain:
				s += shared.StyleAgain.Render(fmt.Sprintf("[1] Again (%s)", interval)) + "  "
			case domain.RatingHard:
				s += shared.StyleHard.Render(fmt.Sprintf("[2] Hard (%s)", interval)) + "  "
			case domain.RatingGood:
				s += shared.StyleGood.Render(fmt.Sprintf("[3] Good (%s)", interval)) + "  "
			case domain.RatingEasy:
				s += shared.StyleEasy.Render(fmt.Sprintf("[4] Easy (%s)", interval)) + "  "
			}
		}
	} else {
		s = shared.StyleAgain.Render("[1] Again") + "  " +
			shared.StyleHard.Render("[2] Hard") + "  " +
			shared.StyleGood.Render("[3] Good") + "  " +
			shared.StyleEasy.Render("[4] Easy")
	}
	return "\n" + s + "\n" + shared.StyleHelp.Render("esc unflip • q quit")
}

func (m Model) renderSessionStats() string {
	if m.session == nil {
		return ""
	}
	remaining := len(m.session.Queue) - m.session.Current

	// Current session counts
	var sa, sh, sg, se int
	for _, l := range m.session.Logs {
		switch l.Rating {
		case domain.RatingAgain:
			sa++
		case domain.RatingHard:
			sh++
		case domain.RatingGood:
			sg++
		case domain.RatingEasy:
			se++
		}
	}

	// Combined: prior sessions + current session
	ta := m.priorStats.again + sa
	th := m.priorStats.hard + sh
	tg := m.priorStats.good + sg
	te := m.priorStats.easy + se
	total := ta + th + tg + te

	stats := shared.StyleSubtle.Render(fmt.Sprintf("Remaining: %d", remaining))
	if total > 0 {
		stats += shared.StyleSubtle.Render("  |  ")
		stats += shared.StyleAgain.Render(fmt.Sprintf("A:%d", ta)) + " "
		stats += shared.StyleHard.Render(fmt.Sprintf("H:%d", th)) + " "
		stats += shared.StyleGood.Render(fmt.Sprintf("G:%d", tg)) + " "
		stats += shared.StyleEasy.Render(fmt.Sprintf("E:%d", te))
		pct := float64(tg+te) / float64(total) * 100
		stats += shared.StyleSubtle.Render(fmt.Sprintf("  |  %.0f%% pass", pct))
		stats += shared.StyleSubtle.Render("  (30d)")
	}
	return stats
}

func formatInterval(days uint64) string {
	if days == 0 {
		return "<1d"
	}
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 365 {
		return fmt.Sprintf("%dmo", days/30)
	}
	return fmt.Sprintf("%dy", days/365)
}
