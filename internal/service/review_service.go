package service

import (
	"context"
	"math/rand/v2"
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
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
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

	// Re-queue "Again" cards ~10 cards later for within-session repetition
	if rating == domain.RatingAgain {
		reinsertAt := session.Current + againRequeueGap
		if reinsertAt > len(session.Queue) {
			reinsertAt = len(session.Queue)
		}
		updated := result.Card
		session.Queue = append(session.Queue[:reinsertAt],
			append([]*domain.Card{&updated}, session.Queue[reinsertAt:]...)...)
	}

	return nil
}

const againRequeueGap = 10

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

// PriorRatingCounts returns aggregate rating counts for a deck over the last N days.
func (s *ReviewService) PriorRatingCounts(ctx context.Context, deckName string, days int) (again, hard, good, easy int) {
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return 0, 0, 0, 0
	}
	since := time.Now().UTC().AddDate(0, 0, -days)
	logs, err := s.store.Reviews.ListReviewLogs(ctx, deck.ID, since)
	if err != nil {
		return 0, 0, 0, 0
	}
	for _, l := range logs {
		switch l.Rating {
		case domain.RatingAgain:
			again++
		case domain.RatingHard:
			hard++
		case domain.RatingGood:
			good++
		case domain.RatingEasy:
			easy++
		}
	}
	return again, hard, good, easy
}

func (s *ReviewService) ReviewStats(ctx context.Context, deckID string, days int) ([]*domain.ReviewLog, error) {
	since := time.Now().UTC().AddDate(0, 0, -days)
	return s.store.Reviews.ListReviewLogs(ctx, deckID, since)
}
