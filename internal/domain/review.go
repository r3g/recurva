package domain

import "time"

type Rating int

const (
	RatingAgain Rating = 1
	RatingHard  Rating = 2
	RatingGood  Rating = 3
	RatingEasy  Rating = 4
)

func (r Rating) String() string {
	switch r {
	case RatingAgain:
		return "Again"
	case RatingHard:
		return "Hard"
	case RatingGood:
		return "Good"
	case RatingEasy:
		return "Easy"
	default:
		return "Unknown"
	}
}

type ReviewLog struct {
	ID            string
	CardID        string
	DeckID        string
	Rating        Rating
	State         State
	ScheduledDays uint64
	ElapsedDays   uint64
	ReviewedAt    time.Time
}

type ReviewSession struct {
	Queue   []*Card
	Current int
	Logs    []ReviewLog
	StartAt time.Time
}

func (rs *ReviewSession) CurrentCard() *Card {
	if rs.Current >= len(rs.Queue) {
		return nil
	}
	return rs.Queue[rs.Current]
}

func (rs *ReviewSession) Done() bool {
	return rs.Current >= len(rs.Queue)
}

func (rs *ReviewSession) Progress() (current, total int) {
	return rs.Current + 1, len(rs.Queue)
}

type SessionSummary struct {
	Total    int
	Again    int
	Hard     int
	Good     int
	Easy     int
	Duration time.Duration
}
