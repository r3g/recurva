package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store"
)

type Store struct {
	mu      sync.RWMutex
	decks   map[string]*domain.Deck
	cards   map[string]*domain.Card
	reviews []*domain.ReviewLog
	tags    map[string]*domain.Tag
}

func New() *store.Store {
	s := &Store{
		decks: make(map[string]*domain.Deck),
		cards: make(map[string]*domain.Card),
		tags:  make(map[string]*domain.Tag),
	}
	return &store.Store{
		Cards:   s,
		Decks:   s,
		Reviews: s,
		Tags:    s,
	}
}

// DeckStore methods

func (s *Store) GetDeck(ctx context.Context, id string) (*domain.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.decks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *Store) GetDeckByName(ctx context.Context, name string) (*domain.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.decks {
		if d.Name == name {
			cp := *d
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Store) ListDecks(ctx context.Context) ([]*domain.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var decks []*domain.Deck
	for _, d := range s.decks {
		cp := *d
		decks = append(decks, &cp)
	}
	return decks, nil
}

func (s *Store) CreateDeck(ctx context.Context, deck *domain.Deck) (*domain.Deck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.decks {
		if d.Name == deck.Name {
			return nil, domain.ErrAlreadyExists
		}
	}
	if deck.ID == "" {
		deck.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	deck.CreatedAt = now
	deck.UpdatedAt = now
	cp := *deck
	s.decks[deck.ID] = &cp
	return deck, nil
}

func (s *Store) DeleteDeck(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.decks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.decks, id)
	return nil
}

func (s *Store) DeckStats(ctx context.Context, deckID string, now time.Time) (*domain.DeckStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.decks[deckID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	stats := &domain.DeckStats{DeckID: deckID, DeckName: d.Name}
	for _, c := range s.cards {
		if c.DeckID != deckID {
			continue
		}
		stats.TotalCards++
		if !c.Due.After(now) {
			stats.DueCards++
		}
		if c.SRS.State == domain.StateNew {
			stats.NewCards++
		}
	}
	return stats, nil
}

// CardStore methods

func (s *Store) GetCard(ctx context.Context, id string) (*domain.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.cards[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (s *Store) ListCards(ctx context.Context, deckID string, dueOnly bool, now time.Time) ([]*domain.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var cards []*domain.Card
	for _, c := range s.cards {
		if c.DeckID != deckID {
			continue
		}
		if dueOnly && c.Due.After(now) {
			continue
		}
		cp := *c
		cards = append(cards, &cp)
	}
	return cards, nil
}

func (s *Store) CreateCard(ctx context.Context, card *domain.Card) (*domain.Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if card.ID == "" {
		card.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	card.CreatedAt = now
	card.UpdatedAt = now
	if card.Due.IsZero() {
		card.Due = now
	}
	cp := *card
	s.cards[card.ID] = &cp
	return card, nil
}

func (s *Store) UpdateCard(ctx context.Context, card *domain.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cards[card.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *card
	s.cards[card.ID] = &cp
	return nil
}

func (s *Store) DeleteCard(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cards[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.cards, id)
	return nil
}

func (s *Store) BulkCreateCards(ctx context.Context, cards []*domain.Card) error {
	for _, c := range cards {
		if _, err := s.CreateCard(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// ReviewStore methods

func (s *Store) CreateReviewLog(ctx context.Context, log *domain.ReviewLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	cp := *log
	s.reviews = append(s.reviews, &cp)
	return nil
}

func (s *Store) ListReviewLogs(ctx context.Context, deckID string, since time.Time) ([]*domain.ReviewLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var logs []*domain.ReviewLog
	for _, l := range s.reviews {
		if !l.ReviewedAt.Before(since) && (deckID == "" || l.DeckID == deckID) {
			cp := *l
			logs = append(logs, &cp)
		}
	}
	return logs, nil
}

// TagStore methods

func (s *Store) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tags []*domain.Tag
	for _, t := range s.tags {
		cp := *t
		tags = append(tags, &cp)
	}
	return tags, nil
}

func (s *Store) GetTagByName(ctx context.Context, name string) (*domain.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tags {
		if t.Name == name {
			cp := *t
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Store) CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tags {
		if t.Name == tag.Name {
			return nil, domain.ErrAlreadyExists
		}
	}
	if tag.ID == "" {
		tag.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	tag.CreatedAt = now
	tag.UpdatedAt = now
	cp := *tag
	s.tags[tag.ID] = &cp
	return tag, nil
}

func (s *Store) RenameTag(ctx context.Context, id, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tags[id]
	if !ok {
		return domain.ErrNotFound
	}
	for _, existing := range s.tags {
		if existing.Name == newName {
			return domain.ErrAlreadyExists
		}
	}
	t.Name = newName
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Store) DeleteTag(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tags[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.tags, id)
	return nil
}
