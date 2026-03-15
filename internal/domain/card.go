package domain

import "time"

type State int

const (
	StateNew        State = 0
	StateLearning   State = 1
	StateReview     State = 2
	StateRelearning State = 3
)

type SRSData struct {
	Stability     float64
	Difficulty    float64
	ElapsedDays   uint64
	ScheduledDays uint64
	Reps          uint32
	Lapses        uint32
	State         State
	LastReview    time.Time
}

type Card struct {
	ID        string
	DeckID    string
	Front     string
	Back      string
	Notes     string
	Tags      []string
	Due       time.Time
	SRS       SRSData
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Deck struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeckStats struct {
	DeckID     string
	DeckName   string
	TotalCards int
	DueCards   int
	NewCards   int
}

type Tag struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
