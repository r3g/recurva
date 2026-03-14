package scheduler

import (
	"time"

	"github.com/r3g/recurva/internal/domain"
)

type ScheduleResult struct {
	Card      domain.Card
	ReviewLog domain.ReviewLog
}

type RatingPreview struct {
	Rating   domain.Rating
	Interval uint64 // days until next review
}

type Preview struct {
	Ratings []RatingPreview
}

type Scheduler interface {
	Schedule(card domain.Card, rating domain.Rating, now time.Time) (ScheduleResult, error)
	Preview(card domain.Card, now time.Time) (Preview, error)
	Name() string
}
