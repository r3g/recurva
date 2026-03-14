package shared

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen identifiers
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenDecks
	ScreenReview
	ScreenResult
)

// SwitchScreenMsg is sent to navigate between screens
type SwitchScreenMsg struct {
	Screen   Screen
	DeckName string
}

// KeyMap holds all key bindings
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
	Quit   key.Binding
	Flip   key.Binding
	Again  key.Binding
	Hard   key.Binding
	Good   key.Binding
	Easy   key.Binding
}

var DefaultKeyMap = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Flip:   key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "flip")),
	Again:  key.NewBinding(key.WithKeys("1", "a"), key.WithHelp("1/a", "again")),
	Hard:   key.NewBinding(key.WithKeys("2", "h"), key.WithHelp("2/h", "hard")),
	Good:   key.NewBinding(key.WithKeys("3", "g"), key.WithHelp("3/g", "good")),
	Easy:   key.NewBinding(key.WithKeys("4", "e"), key.WithHelp("4/e", "easy")),
}

// Matches is a convenience wrapper for key.Matches
func Matches(msg tea.KeyMsg, b key.Binding) bool {
	return key.Matches(msg, b)
}

// Styles
var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#A78BFA")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorAgain     = lipgloss.Color("#FF2D95")
	ColorHard      = lipgloss.Color("#FF6B1A")
	ColorGood      = lipgloss.Color("#39FF14")
	ColorEasy      = lipgloss.Color("#00D4FF")

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StyleSubtle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	StyleFront = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB"))

	StyleBack = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1FAE5"))

	StyleProgress = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleAgain = lipgloss.NewStyle().Foreground(ColorAgain).Bold(true)
	StyleHard  = lipgloss.NewStyle().Foreground(ColorHard).Bold(true)
	StyleGood  = lipgloss.NewStyle().Foreground(ColorGood).Bold(true)
	StyleEasy  = lipgloss.NewStyle().Foreground(ColorEasy).Bold(true)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)
