package domain

import (
	"testing"
	"time"
)

func TestRatingString(t *testing.T) {
	tests := []struct {
		r    Rating
		want string
	}{
		{RatingAgain, "Again"},
		{RatingHard, "Hard"},
		{RatingGood, "Good"},
		{RatingEasy, "Easy"},
		{Rating(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("Rating(%d).String() = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestReviewSession_CurrentCard(t *testing.T) {
	cards := []*Card{
		{ID: "a", Front: "Q1"},
		{ID: "b", Front: "Q2"},
	}
	session := &ReviewSession{Queue: cards}

	// First card
	got := session.CurrentCard()
	if got == nil || got.ID != "a" {
		t.Fatalf("CurrentCard() = %v, want card 'a'", got)
	}

	// Advance past end
	session.Current = 2
	if got := session.CurrentCard(); got != nil {
		t.Fatalf("CurrentCard() past end = %v, want nil", got)
	}
}

func TestReviewSession_Done(t *testing.T) {
	session := &ReviewSession{Queue: make([]*Card, 3)}

	if session.Done() {
		t.Fatal("Done() = true with 3 cards and current=0")
	}

	session.Current = 3
	if !session.Done() {
		t.Fatal("Done() = false with current=len(queue)")
	}
}

func TestReviewSession_Progress(t *testing.T) {
	session := &ReviewSession{Queue: make([]*Card, 5), Current: 2}
	cur, total := session.Progress()
	if cur != 3 || total != 5 {
		t.Errorf("Progress() = (%d, %d), want (3, 5)", cur, total)
	}
}

func TestReviewSession_EmptyQueue(t *testing.T) {
	session := &ReviewSession{}

	if !session.Done() {
		t.Fatal("empty session should be done")
	}
	if c := session.CurrentCard(); c != nil {
		t.Fatalf("empty session CurrentCard() = %v, want nil", c)
	}
	cur, total := session.Progress()
	if cur != 1 || total != 0 {
		t.Errorf("empty session Progress() = (%d, %d), want (1, 0)", cur, total)
	}
}

func TestStateConstants(t *testing.T) {
	// Values must match go-fsrs for clean casting
	if StateNew != 0 || StateLearning != 1 || StateReview != 2 || StateRelearning != 3 {
		t.Fatal("State constants don't match expected values")
	}
}

func TestRatingConstants(t *testing.T) {
	// Values must match go-fsrs for clean casting
	if RatingAgain != 1 || RatingHard != 2 || RatingGood != 3 || RatingEasy != 4 {
		t.Fatal("Rating constants don't match expected values")
	}
}

func TestCardZeroValues(t *testing.T) {
	var c Card
	if c.Due != (time.Time{}) {
		t.Error("zero Card.Due should be zero time")
	}
	if c.SRS.State != StateNew {
		t.Errorf("zero SRSData.State = %d, want StateNew (0)", c.SRS.State)
	}
}
