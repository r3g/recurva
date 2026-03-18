package fsrs

import (
	"time"

	gofsrs "github.com/open-spaced-repetition/go-fsrs/v4"
	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/scheduler"
)

type FSRSScheduler struct {
	f *gofsrs.FSRS
}

func New(params gofsrs.Parameters) *FSRSScheduler {
	return &FSRSScheduler{f: gofsrs.NewFSRS(params)}
}

func NewDefault() *FSRSScheduler {
	return &FSRSScheduler{f: gofsrs.NewFSRS(gofsrs.DefaultParam())}
}

func domainToFSRS(c domain.Card) gofsrs.Card {
	return gofsrs.Card{
		Due:           c.Due,
		Stability:     c.SRS.Stability,
		Difficulty:    c.SRS.Difficulty,
		ElapsedDays:   c.SRS.ElapsedDays,
		ScheduledDays: c.SRS.ScheduledDays,
		Reps:          uint64(c.SRS.Reps),
		Lapses:        uint64(c.SRS.Lapses),
		State:         gofsrs.State(c.SRS.State),
		LastReview:    c.SRS.LastReview,
	}
}

func fsrsToDomain(fc gofsrs.Card, orig domain.Card) domain.Card {
	orig.Due = fc.Due
	orig.SRS = domain.SRSData{
		Stability:     fc.Stability,
		Difficulty:    fc.Difficulty,
		ElapsedDays:   fc.ElapsedDays,
		ScheduledDays: fc.ScheduledDays,
		Reps:          uint32(fc.Reps),
		Lapses:        uint32(fc.Lapses),
		State:         domain.State(fc.State),
		LastReview:    fc.LastReview,
	}
	return orig
}

func (s *FSRSScheduler) Schedule(card domain.Card, rating domain.Rating, now time.Time) (scheduler.ScheduleResult, error) {
	fc := domainToFSRS(card)
	info := s.f.Next(fc, now, gofsrs.Rating(rating))
	updatedCard := fsrsToDomain(info.Card, card)

	log := domain.ReviewLog{
		CardID:        card.ID,
		DeckID:        card.DeckID,
		Rating:        rating,
		State:         domain.State(fc.State),
		ScheduledDays: info.Card.ScheduledDays,
		ElapsedDays:   info.Card.ElapsedDays,
		ReviewedAt:    now,
	}

	return scheduler.ScheduleResult{
		Card:      updatedCard,
		ReviewLog: log,
	}, nil
}

func (s *FSRSScheduler) Preview(card domain.Card, now time.Time) (scheduler.Preview, error) {
	fc := domainToFSRS(card)
	recordLog := s.f.Repeat(fc, now)

	ratings := []domain.Rating{domain.RatingAgain, domain.RatingHard, domain.RatingGood, domain.RatingEasy}
	previews := make([]scheduler.RatingPreview, 0, len(ratings))
	for _, r := range ratings {
		info := recordLog[gofsrs.Rating(r)]
		previews = append(previews, scheduler.RatingPreview{
			Rating:   r,
			Interval: info.Card.Due.Sub(now),
		})
	}

	return scheduler.Preview{Ratings: previews}, nil
}

func (s *FSRSScheduler) Name() string {
	return "FSRS v4"
}
