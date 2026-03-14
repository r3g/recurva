package service

import (
	"testing"
	"time"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/scheduler/fsrs"
	"github.com/r3g/recurva/internal/store/memory"
)

func setupReviewTest(t *testing.T) (*ReviewService, *DeckService, *CardService) {
	t.Helper()
	s := memory.New()
	sched := fsrs.NewDefault()
	deckSvc := NewDeckService(*s)
	cardSvc := NewCardService(*s)
	reviewSvc := NewReviewService(*s, sched)

	if _, err := deckSvc.CreateDeck(ctx(), "Go", ""); err != nil {
		t.Fatalf("setup deck: %v", err)
	}
	return reviewSvc, deckSvc, cardSvc
}

func TestReviewService_FullWorkflow(t *testing.T) {
	reviewSvc, _, cardSvc := setupReviewTest(t)

	if _, err := cardSvc.AddCard(ctx(), "Go", "What is Go?", "A language", "", nil); err != nil {
		t.Fatalf("setup card: %v", err)
	}
	if _, err := cardSvc.AddCard(ctx(), "Go", "What is a goroutine?", "A lightweight thread", "", nil); err != nil {
		t.Fatalf("setup card: %v", err)
	}

	// Start session
	session, err := reviewSvc.StartSession(ctx(), "Go")
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if len(session.Queue) != 2 {
		t.Fatalf("queue = %d, want 2", len(session.Queue))
	}
	if session.Done() {
		t.Fatal("session should not be done")
	}

	// Rate first card
	if err := reviewSvc.Rate(ctx(), session, domain.RatingGood); err != nil {
		t.Fatalf("Rate(Good): %v", err)
	}
	if session.Current != 1 {
		t.Errorf("Current = %d, want 1", session.Current)
	}
	if len(session.Logs) != 1 {
		t.Errorf("Logs = %d, want 1", len(session.Logs))
	}

	// Rate second card
	if err := reviewSvc.Rate(ctx(), session, domain.RatingEasy); err != nil {
		t.Fatalf("Rate(Easy): %v", err)
	}
	if !session.Done() {
		t.Fatal("session should be done after rating all cards")
	}

	// Summary
	summary := reviewSvc.Summary(session)
	if summary.Total != 2 {
		t.Errorf("Total = %d, want 2", summary.Total)
	}
	if summary.Good != 1 {
		t.Errorf("Good = %d, want 1", summary.Good)
	}
	if summary.Easy != 1 {
		t.Errorf("Easy = %d, want 1", summary.Easy)
	}
}

func TestReviewService_StartSession_NoDueCards(t *testing.T) {
	s := memory.New()
	sched := fsrs.NewDefault()
	deckSvc := NewDeckService(*s)
	reviewSvc := NewReviewService(*s, sched)

	deck, err := deckSvc.CreateDeck(ctx(), "Go", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "q", Back: "a",
		Due: time.Now().Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("setup card: %v", err)
	}

	session, err := reviewSvc.StartSession(ctx(), "Go")
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if len(session.Queue) != 0 {
		t.Fatalf("queue = %d, want 0 (no due cards)", len(session.Queue))
	}
	if !session.Done() {
		t.Fatal("empty session should be done")
	}
}

func TestReviewService_StartSession_DeckNotFound(t *testing.T) {
	reviewSvc := NewReviewService(*memory.New(), fsrs.NewDefault())

	_, err := reviewSvc.StartSession(ctx(), "Nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestReviewService_RatePastEnd(t *testing.T) {
	reviewSvc := NewReviewService(*memory.New(), fsrs.NewDefault())

	session := &domain.ReviewSession{}
	err := reviewSvc.Rate(ctx(), session, domain.RatingGood)
	if err == nil {
		t.Fatal("Rate past end should error")
	}
}

func TestReviewService_Preview(t *testing.T) {
	reviewSvc := NewReviewService(*memory.New(), fsrs.NewDefault())

	card := domain.Card{
		ID: "c1", DeckID: "d1",
		Front: "q", Back: "a",
		Due: time.Now(),
		SRS: domain.SRSData{State: domain.StateNew},
	}

	preview, err := reviewSvc.Preview(card)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if len(preview.Ratings) != 4 {
		t.Fatalf("preview ratings = %d, want 4", len(preview.Ratings))
	}
}

func TestReviewService_CardUpdatedAfterRating(t *testing.T) {
	s := memory.New()
	sched := fsrs.NewDefault()
	deckSvc := NewDeckService(*s)
	cardSvc := NewCardService(*s)
	reviewSvc := NewReviewService(*s, sched)

	if _, err := deckSvc.CreateDeck(ctx(), "Go", ""); err != nil {
		t.Fatalf("setup: %v", err)
	}
	card, err := cardSvc.AddCard(ctx(), "Go", "Q", "A", "", nil)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	session, _ := reviewSvc.StartSession(ctx(), "Go")
	if err := reviewSvc.Rate(ctx(), session, domain.RatingGood); err != nil {
		t.Fatalf("Rate: %v", err)
	}

	// Fetch the card from store — should be updated with SRS data
	updated, err := s.Cards.GetCard(ctx(), card.ID)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if updated.SRS.Reps != 1 {
		t.Errorf("Reps = %d, want 1 after rating", updated.SRS.Reps)
	}
	if updated.SRS.Stability <= 0 {
		t.Error("Stability should be > 0 after rating")
	}
}

func TestReviewService_ReviewStats(t *testing.T) {
	reviewSvc, deckSvc, cardSvc := setupReviewTest(t)

	if _, err := cardSvc.AddCard(ctx(), "Go", "Q", "A", "", nil); err != nil {
		t.Fatalf("setup: %v", err)
	}

	session, _ := reviewSvc.StartSession(ctx(), "Go")
	if err := reviewSvc.Rate(ctx(), session, domain.RatingGood); err != nil {
		t.Fatalf("Rate: %v", err)
	}

	deck, err := deckSvc.GetDeckByName(ctx(), "Go")
	if err != nil {
		t.Fatalf("GetDeckByName: %v", err)
	}
	logs, err := reviewSvc.ReviewStats(ctx(), deck.ID, 1)
	if err != nil {
		t.Fatalf("ReviewStats: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs = %d, want 1", len(logs))
	}
}
