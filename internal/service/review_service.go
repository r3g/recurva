package service

import (
	"context"
	"time"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/scheduler"
	"github.com/r3g/recurva/internal/store"
)

func timeNow() time.Time { return time.Now().UTC() }

type ReviewService struct {
	store     store.Store
	scheduler scheduler.Scheduler
}

func NewReviewService(s store.Store, sched scheduler.Scheduler) *ReviewService {
	return &ReviewService{store: s, scheduler: sched}
}

func (s *ReviewService) StartSession(ctx context.Context, deckName string) (*domain.ReviewSession, error) {
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	cards, err := s.store.Cards.ListCards(ctx, deck.ID, true, now)
	if err != nil {
		return nil, err
	}
	return &domain.ReviewSession{
		Queue:   cards,
		StartAt: now,
	}, nil
}

func (s *ReviewService) Rate(ctx context.Context, session *domain.ReviewSession, rating domain.Rating) error {
	card := session.CurrentCard()
	if card == nil {
		return domain.ErrInvalidInput
	}

	result, err := s.scheduler.Schedule(*card, rating, time.Now().UTC())
	if err != nil {
		return err
	}

	if err := s.store.Cards.UpdateCard(ctx, &result.Card); err != nil {
		return err
	}

	log := result.ReviewLog
	if err := s.store.Reviews.CreateReviewLog(ctx, &log); err != nil {
		return err
	}

	session.Queue[session.Current] = &result.Card
	session.Logs = append(session.Logs, log)
	session.Current++
	return nil
}

func (s *ReviewService) Preview(card domain.Card) (scheduler.Preview, error) {
	return s.scheduler.Preview(card, time.Now().UTC())
}

func (s *ReviewService) Summary(session *domain.ReviewSession) domain.SessionSummary {
	summary := domain.SessionSummary{
		Total:    len(session.Logs),
		Duration: time.Since(session.StartAt),
	}
	for _, l := range session.Logs {
		switch l.Rating {
		case domain.RatingAgain:
			summary.Again++
		case domain.RatingHard:
			summary.Hard++
		case domain.RatingGood:
			summary.Good++
		case domain.RatingEasy:
			summary.Easy++
		}
	}
	return summary
}

func (s *ReviewService) ReviewStats(ctx context.Context, deckID string, days int) ([]*domain.ReviewLog, error) {
	since := time.Now().UTC().AddDate(0, 0, -days)
	return s.store.Reviews.ListReviewLogs(ctx, deckID, since)
}
