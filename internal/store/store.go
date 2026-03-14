package store

import (
	"context"
	"time"

	"github.com/r3g/recurva/internal/domain"
)

type CardStore interface {
	GetCard(ctx context.Context, id string) (*domain.Card, error)
	ListCards(ctx context.Context, deckID string, dueOnly bool, now time.Time) ([]*domain.Card, error)
	CreateCard(ctx context.Context, card *domain.Card) (*domain.Card, error)
	UpdateCard(ctx context.Context, card *domain.Card) error
	DeleteCard(ctx context.Context, id string) error
	BulkCreateCards(ctx context.Context, cards []*domain.Card) error
}

type DeckStore interface {
	GetDeck(ctx context.Context, id string) (*domain.Deck, error)
	GetDeckByName(ctx context.Context, name string) (*domain.Deck, error)
	ListDecks(ctx context.Context) ([]*domain.Deck, error)
	CreateDeck(ctx context.Context, deck *domain.Deck) (*domain.Deck, error)
	DeleteDeck(ctx context.Context, id string) error
	DeckStats(ctx context.Context, deckID string, now time.Time) (*domain.DeckStats, error)
}

type ReviewStore interface {
	CreateReviewLog(ctx context.Context, log *domain.ReviewLog) error
	ListReviewLogs(ctx context.Context, deckID string, since time.Time) ([]*domain.ReviewLog, error)
}

type Store struct {
	Cards   CardStore
	Decks   DeckStore
	Reviews ReviewStore
}
