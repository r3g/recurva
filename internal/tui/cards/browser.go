package cards

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/tui/shared"
)

// Browser states
type browserState int

const (
	browserStateLoading browserState = iota
	browserStateList
	browserStateFilter
	browserStateDetail
	browserStateEdit
	browserStateAdd
	browserStateDelete
)

// Key constants to satisfy goconst.
const (
	keyTab   = "tab"
	keyCtrlS = "ctrl+s"
	keyEsc   = "esc"
)

// Messages
type cardsLoadedMsg struct {
	cards []*domain.Card
	err   error
}

type (
	cardSavedMsg   struct{ err error }
	cardDeletedMsg struct{ err error }
)

// BrowserModel is the card browser TUI component.
type BrowserModel struct {
	cardSvc  *service.CardService
	deckName string
	state    browserState

	// Card data
	cards    []*domain.Card
	filtered []*domain.Card
	cursor   int
	offset   int // viewport scroll offset
	width    int // terminal width
	height   int // terminal height for viewport calc

	// Filter
	filterInput textinput.Model
	filterText  string

	// Editor fields (edit + add)
	editFront textinput.Model
	editBack  textarea.Model
	editNotes textinput.Model
	editField field // active field (reuse from editor.go)

	// Status
	err     error
	message string
}

func NewBrowser(svc *service.CardService, deckName string, width, height int) (BrowserModel, tea.Cmd) {
	fi := textinput.New()
	fi.Placeholder = "Search cards..."

	m := BrowserModel{
		cardSvc:     svc,
		deckName:    deckName,
		state:       browserStateLoading,
		filterInput: fi,
		width:       width,
		height:      height,
	}
	return m, m.loadCards()
}

func (m BrowserModel) Init() tea.Cmd {
	return m.loadCards()
}

func (m BrowserModel) loadCards() tea.Cmd {
	return func() tea.Msg {
		cards, err := m.cardSvc.ListCards(context.Background(), m.deckName)
		return cardsLoadedMsg{cards: cards, err: err}
	}
}

func (m BrowserModel) viewportRows() int {
	rows := 20
	if m.height > 0 {
		// Reserve lines for header(2) + footer(3) + padding
		rows = m.height - 8
		if rows < 5 {
			rows = 5
		}
	}
	return rows
}

func (m BrowserModel) activeList() []*domain.Card {
	if m.filterText != "" {
		return m.filtered
	}
	return m.cards
}

func (m *BrowserModel) applyFilter() {
	if m.filterText == "" {
		m.filtered = nil
		return
	}
	q := strings.ToLower(m.filterText)
	m.filtered = nil
	for _, c := range m.cards {
		if strings.Contains(strings.ToLower(c.Front), q) ||
			strings.Contains(strings.ToLower(c.Back), q) {
			m.filtered = append(m.filtered, c)
			continue
		}
		for _, tag := range c.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				m.filtered = append(m.filtered, c)
				break
			}
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m *BrowserModel) initEditor(card *domain.Card) {
	front := textinput.New()
	front.Placeholder = "Front (question)"
	front.CharLimit = 500
	front.Focus()

	back := textarea.New()
	back.Placeholder = "Back (answer)"
	back.CharLimit = 2000

	notes := textinput.New()
	notes.Placeholder = "Notes (optional)"
	notes.CharLimit = 1000

	if card != nil {
		front.SetValue(card.Front)
		back.SetValue(card.Back)
		notes.SetValue(card.Notes)
	}

	m.editFront = front
	m.editBack = back
	m.editNotes = notes
	m.editField = fieldFront
	m.err = nil
	m.message = ""
}

func (m *BrowserModel) cycleField() {
	m.editField = (m.editField + 1) % 3
	m.editFront.Blur()
	m.editBack.Blur()
	m.editNotes.Blur()
	switch m.editField {
	case fieldFront:
		m.editFront.Focus()
	case fieldBack:
		m.editBack.Focus()
	case fieldNotes:
		m.editNotes.Focus()
	}
}

func (m *BrowserModel) ensureCursorVisible() {
	rows := m.viewportRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
}

func (m BrowserModel) selectedCard() *domain.Card {
	list := m.activeList()
	if m.cursor >= 0 && m.cursor < len(list) {
		return list[m.cursor]
	}
	return nil
}

// Update handles all messages for the browser.
func (m BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case cardsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.cards = msg.cards
		m.applyFilter()
		if m.state == browserStateLoading {
			m.state = browserStateList
		}
		return m, nil

	case cardSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.state == browserStateAdd {
			m.message = "Card added!"
			m.initEditor(nil) // clear for another
			return m, nil
		}
		// Edit: go back to detail
		m.message = "Card saved!"
		m.state = browserStateDetail
		return m, m.loadCards() // reload to reflect changes

	case cardDeletedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.message = "Card deleted"
		// Remove from local list
		list := m.activeList()
		if len(list) == 0 || m.cursor >= len(list) {
			m.cursor = 0
		} else if m.cursor > 0 && m.cursor >= len(list)-1 {
			m.cursor--
		}
		m.state = browserStateList
		return m, m.loadCards()

	case tea.KeyMsg:
		// Clear transient messages on any keystroke
		m.message = ""
		m.err = nil

		switch m.state {
		case browserStateLoading:
			// no key handling while loading
		case browserStateList:
			return m.updateList(msg)
		case browserStateFilter:
			return m.updateFilter(msg)
		case browserStateDetail:
			return m.updateDetail(msg)
		case browserStateEdit:
			return m.updateEdit(msg)
		case browserStateAdd:
			return m.updateAdd(msg)
		case browserStateDelete:
			return m.updateDelete(msg)
		}
	}
	return m, nil
}

func (m BrowserModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	list := m.activeList()
	switch {
	case shared.Matches(msg, shared.DefaultKeyMap.Up):
		if m.cursor > 0 {
			m.cursor--
			m.ensureCursorVisible()
		}
	case shared.Matches(msg, shared.DefaultKeyMap.Down):
		if m.cursor < len(list)-1 {
			m.cursor++
			m.ensureCursorVisible()
		}
	case shared.Matches(msg, shared.DefaultKeyMap.Select):
		if m.selectedCard() != nil {
			m.state = browserStateDetail
		}
	case msg.String() == "/":
		m.state = browserStateFilter
		m.filterInput.SetValue(m.filterText)
		m.filterInput.Focus()
		return m, textinput.Blink
	case msg.String() == "a":
		m.state = browserStateAdd
		m.initEditor(nil)
		return m, textinput.Blink
	case shared.Matches(msg, shared.DefaultKeyMap.Back):
		if m.filterText != "" {
			m.filterText = ""
			m.filtered = nil
			m.cursor = 0
			m.offset = 0
			return m, nil
		}
		return m, func() tea.Msg {
			return shared.SwitchScreenMsg{Screen: shared.ScreenDecks}
		}
	case shared.Matches(msg, shared.DefaultKeyMap.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m BrowserModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		m.filterText = ""
		m.filtered = nil
		m.cursor = 0
		m.offset = 0
		m.state = browserStateList
		m.filterInput.Blur()
		return m, nil
	case "enter":
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		m.state = browserStateList
		m.filterInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		// Live filter
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, cmd
	}
}

func (m BrowserModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "e":
		card := m.selectedCard()
		if card != nil {
			m.initEditor(card)
			m.state = browserStateEdit
			return m, textinput.Blink
		}
	case "d":
		if m.selectedCard() != nil {
			m.state = browserStateDelete
		}
	case keyEsc:
		m.state = browserStateList
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m BrowserModel) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyTab:
		m.cycleField()
		return m, nil
	case keyCtrlS:
		card := m.selectedCard()
		if card == nil {
			return m, nil
		}
		front := strings.TrimSpace(m.editFront.Value())
		back := strings.TrimSpace(m.editBack.Value())
		if front == "" {
			m.err = fmt.Errorf("front cannot be empty")
			return m, nil
		}
		if back == "" {
			m.err = fmt.Errorf("back cannot be empty")
			return m, nil
		}
		if len(front) > 500 {
			m.err = fmt.Errorf("front too long (max 500)")
			return m, nil
		}
		if len(back) > 2000 {
			m.err = fmt.Errorf("back too long (max 2000)")
			return m, nil
		}
		notes := strings.TrimSpace(m.editNotes.Value())
		if len(notes) > 1000 {
			m.err = fmt.Errorf("notes too long (max 1000)")
			return m, nil
		}
		// Copy card, modify only content fields
		updated := *card
		updated.Front = front
		updated.Back = back
		updated.Notes = notes
		return m, func() tea.Msg {
			err := m.cardSvc.UpdateCard(context.Background(), &updated)
			return cardSavedMsg{err: err}
		}
	case keyEsc:
		m.state = browserStateDetail
		return m, nil
	default:
		return m.updateEditorFields(msg)
	}
}

func (m BrowserModel) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyTab:
		m.cycleField()
		return m, nil
	case keyCtrlS:
		front := strings.TrimSpace(m.editFront.Value())
		back := strings.TrimSpace(m.editBack.Value())
		if front == "" {
			m.err = fmt.Errorf("front cannot be empty")
			return m, nil
		}
		if back == "" {
			m.err = fmt.Errorf("back cannot be empty")
			return m, nil
		}
		if len(front) > 500 {
			m.err = fmt.Errorf("front too long (max 500)")
			return m, nil
		}
		if len(back) > 2000 {
			m.err = fmt.Errorf("back too long (max 2000)")
			return m, nil
		}
		notes := strings.TrimSpace(m.editNotes.Value())
		if len(notes) > 1000 {
			m.err = fmt.Errorf("notes too long (max 1000)")
			return m, nil
		}
		deckName := m.deckName
		return m, func() tea.Msg {
			_, err := m.cardSvc.AddCard(context.Background(), deckName, front, back, notes, nil)
			return cardSavedMsg{err: err}
		}
	case keyEsc:
		m.state = browserStateList
		return m, m.loadCards()
	default:
		return m.updateEditorFields(msg)
	}
}

func (m BrowserModel) updateEditorFields(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.editFront, cmd = m.editFront.Update(msg)
	cmds = append(cmds, cmd)
	m.editBack, cmd = m.editBack.Update(msg)
	cmds = append(cmds, cmd)
	m.editNotes, cmd = m.editNotes.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m BrowserModel) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		card := m.selectedCard()
		if card == nil {
			m.state = browserStateList
			return m, nil
		}
		id := card.ID
		return m, func() tea.Msg {
			err := m.cardSvc.DeleteCard(context.Background(), id)
			return cardDeletedMsg{err: err}
		}
	case "n", keyEsc:
		m.state = browserStateDetail
	}
	return m, nil
}

// View renders the browser.
func (m BrowserModel) View() string {
	switch m.state {
	case browserStateLoading:
		return shared.StyleSubtle.Render("Loading cards...")
	case browserStateList, browserStateFilter:
		return m.viewList()
	case browserStateDetail:
		return m.viewDetail()
	case browserStateEdit:
		return m.viewEditor("Edit Card")
	case browserStateAdd:
		return m.viewEditor("Add Card")
	case browserStateDelete:
		return m.viewDelete()
	}
	return ""
}

func (m BrowserModel) viewList() string {
	list := m.activeList()
	s := shared.StyleTitle.Render("Cards") + " "
	s += shared.StyleSubtle.Render(fmt.Sprintf("Deck: %s  (%d cards)", m.deckName, len(list)))
	s += "\n"

	if m.state == browserStateFilter {
		s += "\n" + m.filterInput.View() + "\n"
	} else if m.filterText != "" {
		s += shared.StyleSubtle.Render(fmt.Sprintf("  Filter: %q", m.filterText)) + "\n"
	}
	s += "\n"

	// Layout: cursor(2) + front + gap(2) + definition(fill) + gap(2) + due(10)
	avail := m.width - 4 // app padding (2 each side)
	if avail < 60 {
		avail = 60
	}
	frontWidth := 22
	dueWidth := 10
	defWidth := avail - 2 - frontWidth - 2 - 2 - dueWidth // cursor + gaps
	if defWidth < 20 {
		defWidth = 20
	}

	colFront := lipgloss.NewStyle().Width(frontWidth)
	colDef := lipgloss.NewStyle().Width(defWidth)
	colDue := lipgloss.NewStyle().Width(dueWidth).Align(lipgloss.Right)

	if len(list) == 0 {
		if m.filterText != "" {
			s += shared.StyleSubtle.Render("No cards match filter.")
		} else {
			s += shared.StyleSubtle.Render("No cards in this deck.")
		}
		s += "\n"
	} else {
		rows := m.viewportRows()
		end := m.offset + rows
		if end > len(list) {
			end = len(list)
		}
		for i := m.offset; i < end; i++ {
			c := list[i]
			cursor := "  "
			style := shared.StyleSubtle
			if i == m.cursor {
				cursor = "▶ "
				style = shared.StyleSelected
			}

			front := truncate(c.Front, frontWidth)
			def := truncate(c.Back, defWidth)
			due := c.Due.Format("2006-01-02")

			line := cursor +
				colFront.Render(style.Render(front)) + "  " +
				colDef.Render(shared.StyleSubtle.Render(def)) + "  " +
				colDue.Render(shared.StyleSubtle.Render(due))
			s += line + "\n"
		}
	}

	if m.message != "" {
		s += "\n" + shared.StyleGood.Render(m.message) + "\n"
	}
	if m.err != nil {
		s += "\n" + shared.StyleAgain.Render("Error: "+m.err.Error()) + "\n"
	}

	if m.state == browserStateFilter {
		s += "\n" + shared.StyleHelp.Render("type to filter • enter accept • esc clear")
	} else {
		help := "↑/↓ navigate • enter detail • / filter • a add • esc back"
		if m.filterText != "" {
			help = "↑/↓ navigate • enter detail • / filter • a add • esc clear filter"
		}
		s += "\n" + shared.StyleHelp.Render(help)
	}
	return s
}

func (m BrowserModel) viewDetail() string {
	card := m.selectedCard()
	if card == nil {
		return shared.StyleSubtle.Render("No card selected")
	}

	s := shared.StyleTitle.Render("Card Detail") + "\n\n"

	// Card content in a box
	widgetWidth := 60
	innerWidth := widgetWidth - 2
	contentWidth := innerWidth - 4
	cardStyle := shared.StyleCard.Width(innerWidth)

	centeredFront := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).
		Bold(true).Foreground(shared.ColorFront).Render(card.Front)

	divider := strings.Repeat("─", contentWidth)
	content := centeredFront + "\n\n" +
		shared.StyleSubtle.Render(divider) + "\n\n" +
		shared.StyleBack.Render(card.Back)
	if card.Notes != "" {
		content += "\n\n" + shared.StyleSubtle.Render("Notes: "+card.Notes)
	}
	s += cardStyle.Render(content) + "\n\n"

	// Metadata
	tags := "(none)"
	if len(card.Tags) > 0 {
		tags = strings.Join(card.Tags, ", ")
	}
	s += shared.StyleSubtle.Render(fmt.Sprintf("Tags:    %s", tags)) + "\n"
	s += shared.StyleSubtle.Render(fmt.Sprintf("State:   %s", stateLabel(card.SRS.State))) + "\n"
	s += shared.StyleSubtle.Render(fmt.Sprintf("Due:     %s", card.Due.Format("2006-01-02"))) + "\n"
	s += shared.StyleSubtle.Render(fmt.Sprintf("Reps:    %d", card.SRS.Reps)) + "\n"
	s += shared.StyleSubtle.Render(fmt.Sprintf("Lapses:  %d", card.SRS.Lapses)) + "\n"

	if m.message != "" {
		s += "\n" + shared.StyleGood.Render(m.message) + "\n"
	}

	s += "\n" + shared.StyleHelp.Render("e edit • d delete • esc back to list")
	return s
}

func (m BrowserModel) viewEditor(title string) string {
	s := shared.StyleTitle.Render(title) + " "
	s += shared.StyleSubtle.Render("Deck: "+m.deckName) + "\n\n"

	s += "Front:\n" + m.editFront.View() + "\n\n"
	s += "Back:\n" + m.editBack.View() + "\n\n"
	s += "Notes:\n" + m.editNotes.View() + "\n\n"

	if m.message != "" {
		s += shared.StyleGood.Render(m.message) + "\n"
	}
	if m.err != nil {
		s += shared.StyleAgain.Render("Error: "+m.err.Error()) + "\n"
	}

	s += shared.StyleHelp.Render("tab switch fields • ctrl+s save • esc cancel")
	return s
}

func (m BrowserModel) viewDelete() string {
	card := m.selectedCard()
	if card == nil {
		return ""
	}
	s := shared.StyleTitle.Render("Delete Card") + "\n\n"
	s += shared.StyleFront.Render(card.Front) + "\n\n"
	s += shared.StyleAgain.Render("Are you sure you want to delete this card?") + "\n\n"
	s += shared.StyleHelp.Render("y confirm • n/esc cancel")
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func stateLabel(state domain.State) string {
	switch state {
	case domain.StateNew:
		return "New"
	case domain.StateLearning:
		return "Learning"
	case domain.StateReview:
		return "Review"
	case domain.StateRelearning:
		return "Relearn"
	default:
		return "Unknown"
	}
}
