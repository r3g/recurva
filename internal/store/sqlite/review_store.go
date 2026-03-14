package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/r3g/recurva/internal/domain"
)

type ReviewStore struct {
	db *DB
}

func NewReviewStore(db *DB) *ReviewStore {
	return &ReviewStore{db: db}
}

func (s *ReviewStore) CreateReviewLog(ctx context.Context, log *domain.ReviewLog) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO review_logs(id, card_id, deck_id, rating, state, scheduled_days, elapsed_days, reviewed_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		log.ID, log.CardID, log.DeckID, log.Rating, log.State,
		log.ScheduledDays, log.ElapsedDays, log.ReviewedAt,
	)
	return err
}

func (s *ReviewStore) ListReviewLogs(ctx context.Context, deckID string, since time.Time) ([]*domain.ReviewLog, error) {
	query := `SELECT id, card_id, deck_id, rating, state, scheduled_days, elapsed_days, reviewed_at
	          FROM review_logs WHERE reviewed_at >= ?`
	args := []interface{}{since}
	if deckID != "" {
		query += ` AND deck_id = ?`
		args = append(args, deckID)
	}
	query += ` ORDER BY reviewed_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.ReviewLog
	for rows.Next() {
		var l domain.ReviewLog
		if err := rows.Scan(&l.ID, &l.CardID, &l.DeckID, &l.Rating, &l.State,
			&l.ScheduledDays, &l.ElapsedDays, &l.ReviewedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}
