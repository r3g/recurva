package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/r3g/recurva/internal/domain"
)

type TagStore struct {
	db *DB
}

func NewTagStore(db *DB) *TagStore {
	return &TagStore{db: db}
}

func (s *TagStore) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, created_at, updated_at FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, &t)
	}
	return tags, rows.Err()
}

func (s *TagStore) GetTagByName(ctx context.Context, name string) (*domain.Tag, error) {
	var t domain.Tag
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at FROM tags WHERE name = ?`, name,
	).Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TagStore) CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	if tag.ID == "" {
		tag.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tags(id, name, created_at, updated_at) VALUES(?,?,?,?)`,
		tag.ID, tag.Name, tag.CreatedAt, tag.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, err
	}
	return tag, nil
}

func (s *TagStore) RenameTag(ctx context.Context, id, newName string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE tags SET name = ?, updated_at = ? WHERE id = ?`,
		newName, time.Now().UTC(), id,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *TagStore) DeleteTag(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
