package fsrs

import (
	"testing"
	"time"

	"github.com/r3g/recurva/internal/domain"
)

func newCard() domain.Card {
	return domain.Card{
		ID:     "test-card",
		DeckID: "test-deck",
		Front:  "What is Go?",
		Back:   "A programming language",
		Due:    time.Now(),
		SRS:    domain.SRSData{State: domain.StateNew},
	}
}

func TestSchedule_NewCard(t *testing.T) {
	s := NewDefault()
	card := newCard()
	now := time.Now()

	result, err := s.Schedule(card, domain.RatingGood, now)
	if err != nil {
		t.Fatalf("Schedule() error: %v", err)
	}

	// Card should advance from New state
	if result.Card.SRS.Reps != 1 {
		t.Errorf("Reps = %d, want 1", result.Card.SRS.Reps)
	}
	if result.Card.SRS.Stability <= 0 {
		t.Error("Stability should be > 0 after scheduling")
	}
	if result.Card.SRS.Difficulty <= 0 {
		t.Error("Difficulty should be > 0 after scheduling")
	}

	// ReviewLog should capture the rating
	if result.ReviewLog.Rating != domain.RatingGood {
		t.Errorf("ReviewLog.Rating = %v, want Good", result.ReviewLog.Rating)
	}
	if result.ReviewLog.CardID != "test-card" {
		t.Errorf("ReviewLog.CardID = %q, want %q", result.ReviewLog.CardID, "test-card")
	}
	if result.ReviewLog.State != domain.StateNew {
		t.Errorf("ReviewLog.State = %d, want StateNew (captures pre-review state)", result.ReviewLog.State)
	}
}

func TestSchedule_AllRatings(t *testing.T) {
	s := NewDefault()
	now := time.Now()

	ratings := []domain.Rating{
		domain.RatingAgain,
		domain.RatingHard,
		domain.RatingGood,
		domain.RatingEasy,
	}

	for _, r := range ratings {
		card := newCard()
		result, err := s.Schedule(card, r, now)
		if err != nil {
			t.Fatalf("Schedule(rating=%v) error: %v", r, err)
		}
		if result.Card.ID != card.ID {
			t.Errorf("Schedule(rating=%v) changed card ID", r)
		}
	}
}

func TestSchedule_EasyHasLongerInterval(t *testing.T) {
	s := NewDefault()
	now := time.Now()

	// Schedule a card that has been reviewed once (in Review state)
	baseCard := newCard()
	good, _ := s.Schedule(baseCard, domain.RatingGood, now)

	// Now schedule the reviewed card again with Good vs Easy
	goodResult, _ := s.Schedule(good.Card, domain.RatingGood, now.Add(24*time.Hour))
	easyResult, _ := s.Schedule(good.Card, domain.RatingEasy, now.Add(24*time.Hour))

	if easyResult.Card.SRS.ScheduledDays < goodResult.Card.SRS.ScheduledDays {
		t.Errorf("Easy interval (%d) should be >= Good interval (%d)",
			easyResult.Card.SRS.ScheduledDays, goodResult.Card.SRS.ScheduledDays)
	}
}

func TestPreview(t *testing.T) {
	s := NewDefault()
	card := newCard()
	now := time.Now()

	preview, err := s.Preview(card, now)
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}

	if len(preview.Ratings) != 4 {
		t.Fatalf("Preview() returned %d ratings, want 4", len(preview.Ratings))
	}

	// Verify all four ratings are present in order
	expected := []domain.Rating{
		domain.RatingAgain,
		domain.RatingHard,
		domain.RatingGood,
		domain.RatingEasy,
	}
	for i, rp := range preview.Ratings {
		if rp.Rating != expected[i] {
			t.Errorf("preview.Ratings[%d].Rating = %v, want %v", i, rp.Rating, expected[i])
		}
	}
}

func TestName(t *testing.T) {
	s := NewDefault()
	if got := s.Name(); got != "FSRS v4" {
		t.Errorf("Name() = %q, want %q", got, "FSRS v4")
	}
}
