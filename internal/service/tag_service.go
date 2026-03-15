package service

import (
	"context"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store"
)

type TagService struct {
	store store.Store
}

func NewTagService(s store.Store) *TagService {
	return &TagService{store: s}
}

func (s *TagService) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	return s.store.Tags.ListTags(ctx)
}

func (s *TagService) AddTag(ctx context.Context, name string) (*domain.Tag, error) {
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.Tags.CreateTag(ctx, &domain.Tag{Name: name})
}

func (s *TagService) RenameTag(ctx context.Context, oldName, newName string) error {
	if oldName == "" || newName == "" {
		return domain.ErrInvalidInput
	}
	tag, err := s.store.Tags.GetTagByName(ctx, oldName)
	if err != nil {
		return err
	}
	return s.store.Tags.RenameTag(ctx, tag.ID, newName)
}

func (s *TagService) DeleteTag(ctx context.Context, name string) error {
	tag, err := s.store.Tags.GetTagByName(ctx, name)
	if err != nil {
		return err
	}
	return s.store.Tags.DeleteTag(ctx, tag.ID)
}
