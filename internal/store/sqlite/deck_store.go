package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/r3g/recurva/internal/domain"
)

type DeckStore struct {
	db *DB
}

func NewDeckStore(db *DB) *DeckStore {
	return &DeckStore{db: db}
}

func (s *DeckStore) GetDeck(ctx context.Context, id string) (*domain.Deck, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM decks WHERE id = ?`, id)
	return scanDeck(row)
}

func (s *DeckStore) GetDeckByName(ctx context.Context, name string) (*domain.Deck, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM decks WHERE name = ?`, name)
	return scanDeck(row)
}

func scanDeck(row *sql.Row) (*domain.Deck, error) {
	var d domain.Deck
	err := row.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *DeckStore) ListDecks(ctx context.Context) ([]*domain.Deck, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM decks ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []*domain.Deck
	for rows.Next() {
		var d domain.Deck
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		decks = append(decks, &d)
	}
	return decks, rows.Err()
}

func (s *DeckStore) CreateDeck(ctx context.Context, deck *domain.Deck) (*domain.Deck, error) {
	if deck.ID == "" {
		deck.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	deck.CreatedAt = now
	deck.UpdatedAt = now

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO decks(id, name, description, created_at, updated_at) VALUES(?,?,?,?,?)`,
		deck.ID, deck.Name, deck.Description, deck.CreatedAt, deck.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, err
	}
	return deck, nil
}

func (s *DeckStore) DeleteDeck(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM decks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *DeckStore) DeckStats(ctx context.Context, deckID string, now time.Time) (*domain.DeckStats, error) {
	var stats domain.DeckStats
	stats.DeckID = deckID

	err := s.db.QueryRowContext(ctx, `SELECT name FROM decks WHERE id = ?`, deckID).Scan(&stats.DeckName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN due <= ? THEN 1 ELSE 0 END), 0) as due,
			COALESCE(SUM(CASE WHEN state = 0 THEN 1 ELSE 0 END), 0) as new_cards
		FROM cards WHERE deck_id = ?
	`, now, deckID).Scan(&stats.TotalCards, &stats.DueCards, &stats.NewCards)

	return &stats, err
}

func isUniqueConstraint(err error) bool {
	return err != nil && (fmt.Sprintf("%v", err) == "UNIQUE constraint failed: decks.name" ||
		contains(err.Error(), "UNIQUE constraint failed"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
