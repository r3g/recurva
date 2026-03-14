package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/r3g/recurva/internal/domain"
)

type CardStore struct {
	db *DB
}

func NewCardStore(db *DB) *CardStore {
	return &CardStore{db: db}
}

func (s *CardStore) GetCard(ctx context.Context, id string) (*domain.Card, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, deck_id, front, back, notes, tags, due,
		       stability, difficulty, elapsed_days, scheduled_days,
		       reps, lapses, state, last_review, created_at, updated_at
		FROM cards WHERE id = ?`, id)
	c, err := scanCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return c, err
}

func (s *CardStore) ListCards(ctx context.Context, deckID string, dueOnly bool, now time.Time) ([]*domain.Card, error) {
	query := `
		SELECT id, deck_id, front, back, notes, tags, due,
		       stability, difficulty, elapsed_days, scheduled_days,
		       reps, lapses, state, last_review, created_at, updated_at
		FROM cards WHERE deck_id = ?`
	args := []interface{}{deckID}
	if dueOnly {
		query += ` AND due <= ?`
		args = append(args, now)
	}
	query += ` ORDER BY due`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		c, err := scanCardRow(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (s *CardStore) CreateCard(ctx context.Context, card *domain.Card) (*domain.Card, error) {
	if card.ID == "" {
		card.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	card.CreatedAt = now
	card.UpdatedAt = now
	if card.Due.IsZero() {
		card.Due = now
	}

	tags, _ := json.Marshal(card.Tags)
	var lastReview *time.Time
	if !card.SRS.LastReview.IsZero() {
		t := card.SRS.LastReview
		lastReview = &t
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cards(id, deck_id, front, back, notes, tags, due,
		                  stability, difficulty, elapsed_days, scheduled_days,
		                  reps, lapses, state, last_review, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		card.ID, card.DeckID, card.Front, card.Back, card.Notes, string(tags), card.Due,
		card.SRS.Stability, card.SRS.Difficulty, card.SRS.ElapsedDays, card.SRS.ScheduledDays,
		card.SRS.Reps, card.SRS.Lapses, card.SRS.State, lastReview,
		card.CreatedAt, card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *CardStore) UpdateCard(ctx context.Context, card *domain.Card) error {
	card.UpdatedAt = time.Now().UTC()
	tags, _ := json.Marshal(card.Tags)
	var lastReview *time.Time
	if !card.SRS.LastReview.IsZero() {
		t := card.SRS.LastReview
		lastReview = &t
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE cards SET
		    deck_id=?, front=?, back=?, notes=?, tags=?, due=?,
		    stability=?, difficulty=?, elapsed_days=?, scheduled_days=?,
		    reps=?, lapses=?, state=?, last_review=?, updated_at=?
		WHERE id=?`,
		card.DeckID, card.Front, card.Back, card.Notes, string(tags), card.Due,
		card.SRS.Stability, card.SRS.Difficulty, card.SRS.ElapsedDays, card.SRS.ScheduledDays,
		card.SRS.Reps, card.SRS.Lapses, card.SRS.State, lastReview, card.UpdatedAt,
		card.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *CardStore) DeleteCard(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM cards WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *CardStore) BulkCreateCards(ctx context.Context, cards []*domain.Card) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO cards(id, deck_id, front, back, notes, tags, due,
		                  stability, difficulty, elapsed_days, scheduled_days,
		                  reps, lapses, state, last_review, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, card := range cards {
		if card.ID == "" {
			card.ID = uuid.NewString()
		}
		card.CreatedAt = now
		card.UpdatedAt = now
		if card.Due.IsZero() {
			card.Due = now
		}
		tags, _ := json.Marshal(card.Tags)
		var lastReview *time.Time
		if !card.SRS.LastReview.IsZero() {
			t := card.SRS.LastReview
			lastReview = &t
		}
		_, err := stmt.ExecContext(ctx,
			card.ID, card.DeckID, card.Front, card.Back, card.Notes, string(tags), card.Due,
			card.SRS.Stability, card.SRS.Difficulty, card.SRS.ElapsedDays, card.SRS.ScheduledDays,
			card.SRS.Reps, card.SRS.Lapses, card.SRS.State, lastReview,
			card.CreatedAt, card.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func scanCard(row *sql.Row) (*domain.Card, error) {
	var c domain.Card
	var tagsJSON string
	var lastReview *time.Time
	err := row.Scan(
		&c.ID, &c.DeckID, &c.Front, &c.Back, &c.Notes, &tagsJSON, &c.Due,
		&c.SRS.Stability, &c.SRS.Difficulty, &c.SRS.ElapsedDays, &c.SRS.ScheduledDays,
		&c.SRS.Reps, &c.SRS.Lapses, &c.SRS.State, &lastReview,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
	if lastReview != nil {
		c.SRS.LastReview = *lastReview
	}
	return &c, nil
}

func scanCardRow(rows *sql.Rows) (*domain.Card, error) {
	var c domain.Card
	var tagsJSON string
	var lastReview *time.Time
	err := rows.Scan(
		&c.ID, &c.DeckID, &c.Front, &c.Back, &c.Notes, &tagsJSON, &c.Due,
		&c.SRS.Stability, &c.SRS.Difficulty, &c.SRS.ElapsedDays, &c.SRS.ScheduledDays,
		&c.SRS.Reps, &c.SRS.Lapses, &c.SRS.State, &lastReview,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
	if lastReview != nil {
		c.SRS.LastReview = *lastReview
	}
	return &c, nil
}
