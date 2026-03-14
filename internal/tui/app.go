package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/tui/decks"
	"github.com/r3g/recurva/internal/tui/menu"
	"github.com/r3g/recurva/internal/tui/review"
	"github.com/r3g/recurva/internal/tui/shared"
)

// Re-export screen types from shared
type Screen = shared.Screen

const (
	ScreenMenu   = shared.ScreenMenu
	ScreenDecks  = shared.ScreenDecks
	ScreenReview = shared.ScreenReview
	ScreenResult = shared.ScreenResult
)

// SwitchScreenMsg re-exported from shared
type SwitchScreenMsg = shared.SwitchScreenMsg

type Services struct {
	Decks   *service.DeckService
	Cards   *service.CardService
	Reviews *service.ReviewService
}

type App struct {
	services Services
	screen   Screen
	width    int
	height   int
	menu     menu.Model
	decks    decks.Model
	review   review.Model
	result   review.ResultModel
}

func NewApp(svc Services) *App {
	return &App{
		services: svc,
		screen:   ScreenMenu,
		menu:     menu.New(),
		decks:    decks.New(svc.Decks),
	}
}

func (a *App) Init() tea.Cmd {
	return a.menu.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case shared.SwitchScreenMsg:
		return a.handleSwitch(msg)
	}

	switch a.screen {
	case ScreenMenu:
		m, cmd := a.menu.Update(msg)
		a.menu = m.(menu.Model)
		return a, cmd
	case ScreenDecks:
		m, cmd := a.decks.Update(msg)
		a.decks = m.(decks.Model)
		return a, cmd
	case ScreenReview:
		m, cmd := a.review.Update(msg)
		a.review = m.(review.Model)
		return a, cmd
	case ScreenResult:
		m, cmd := a.result.Update(msg)
		a.result = m.(review.ResultModel)
		return a, cmd
	}
	return a, nil
}

func (a *App) handleSwitch(msg shared.SwitchScreenMsg) (tea.Model, tea.Cmd) {
	switch msg.Screen {
	case ScreenDecks:
		a.screen = ScreenDecks
		return a, a.decks.Init()
	case ScreenReview:
		a.screen = ScreenReview
		m, cmd := review.New(a.services.Reviews, msg.DeckName)
		a.review = m
		return a, cmd
	case ScreenResult:
		a.screen = ScreenResult
		a.result = review.NewResult(a.review.Summary())
		return a, a.result.Init()
	case ScreenMenu:
		a.screen = ScreenMenu
		return a, nil
	}
	return a, nil
}

func (a *App) View() string {
	var content string
	switch a.screen {
	case ScreenMenu:
		content = a.menu.View()
	case ScreenDecks:
		content = a.decks.View()
	case ScreenReview:
		content = a.review.View()
	case ScreenResult:
		content = a.result.View()
	}
	if a.width == 0 || a.height == 0 {
		return content
	}
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}
