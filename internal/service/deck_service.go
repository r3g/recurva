package service

import (
	"context"
	"time"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store"
)

type DeckService struct {
	store store.Store
}

func NewDeckService(s store.Store) *DeckService {
	return &DeckService{store: s}
}

func (s *DeckService) CreateDeck(ctx context.Context, name, description string) (*domain.Deck, error) {
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.Decks.CreateDeck(ctx, &domain.Deck{Name: name, Description: description})
}

func (s *DeckService) ListDecks(ctx context.Context) ([]*domain.Deck, error) {
	return s.store.Decks.ListDecks(ctx)
}

func (s *DeckService) DeleteDeck(ctx context.Context, name string) error {
	d, err := s.store.Decks.GetDeckByName(ctx, name)
	if err != nil {
		return err
	}
	return s.store.Decks.DeleteDeck(ctx, d.ID)
}

func (s *DeckService) GetDeckByName(ctx context.Context, name string) (*domain.Deck, error) {
	return s.store.Decks.GetDeckByName(ctx, name)
}

func (s *DeckService) DeckStats(ctx context.Context, deckID string) (*domain.DeckStats, error) {
	return s.store.Decks.DeckStats(ctx, deckID, time.Now())
}

func (s *DeckService) AllDeckStats(ctx context.Context) ([]*domain.DeckStats, error) {
	decks, err := s.store.Decks.ListDecks(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var result []*domain.DeckStats
	for _, d := range decks {
		stats, err := s.store.Decks.DeckStats(ctx, d.ID, now)
		if err != nil {
			return nil, err
		}
		result = append(result, stats)
	}
	return result, nil
}
