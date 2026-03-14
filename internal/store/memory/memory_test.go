package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/r3g/recurva/internal/domain"
)

func ctx() context.Context { return context.Background() }

func mustCreateDeck(t *testing.T, s *domain.Deck, store interface {
	CreateDeck(context.Context, *domain.Deck) (*domain.Deck, error)
},
) *domain.Deck {
	t.Helper()
	d, err := store.CreateDeck(ctx(), s)
	if err != nil {
		t.Fatalf("setup CreateDeck: %v", err)
	}
	return d
}

// --- Deck tests ---

func TestDeckCRUD(t *testing.T) {
	s := New()

	// Create
	deck, err := s.Decks.CreateDeck(ctx(), &domain.Deck{Name: "Go", Description: "Go lang"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	if deck.ID == "" {
		t.Fatal("CreateDeck did not assign ID")
	}
	if deck.CreatedAt.IsZero() {
		t.Fatal("CreateDeck did not set CreatedAt")
	}

	// Get by ID
	got, err := s.Decks.GetDeck(ctx(), deck.ID)
	if err != nil {
		t.Fatalf("GetDeck: %v", err)
	}
	if got.Name != "Go" {
		t.Errorf("GetDeck name = %q, want %q", got.Name, "Go")
	}

	// Get by name
	got, err = s.Decks.GetDeckByName(ctx(), "Go")
	if err != nil {
		t.Fatalf("GetDeckByName: %v", err)
	}
	if got.ID != deck.ID {
		t.Errorf("GetDeckByName ID mismatch")
	}

	// List
	decks, err := s.Decks.ListDecks(ctx())
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("ListDecks len = %d, want 1", len(decks))
	}

	// Delete
	if err := s.Decks.DeleteDeck(ctx(), deck.ID); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}
	decks, _ = s.Decks.ListDecks(ctx())
	if len(decks) != 0 {
		t.Fatalf("ListDecks after delete len = %d, want 0", len(decks))
	}
}

func TestDeckDuplicateName(t *testing.T) {
	s := New()
	mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	_, err := s.Decks.CreateDeck(ctx(), &domain.Deck{Name: "Go"})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("duplicate deck error = %v, want ErrAlreadyExists", err)
	}
}

func TestDeckNotFound(t *testing.T) {
	s := New()

	_, err := s.Decks.GetDeck(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetDeck error = %v, want ErrNotFound", err)
	}

	_, err = s.Decks.GetDeckByName(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetDeckByName error = %v, want ErrNotFound", err)
	}

	err = s.Decks.DeleteDeck(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteDeck error = %v, want ErrNotFound", err)
	}
}

// --- Card tests ---

func TestCardCRUD(t *testing.T) {
	s := New()
	deck := mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	// Create
	card, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID,
		Front:  "What is Go?",
		Back:   "A language",
	})
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if card.ID == "" {
		t.Fatal("CreateCard did not assign ID")
	}
	if card.Due.IsZero() {
		t.Fatal("CreateCard did not set default Due")
	}

	// Get
	got, err := s.Cards.GetCard(ctx(), card.ID)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if got.Front != "What is Go?" {
		t.Errorf("GetCard front = %q, want %q", got.Front, "What is Go?")
	}

	// Update
	got.Back = "A compiled language"
	if err := s.Cards.UpdateCard(ctx(), got); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	got2, _ := s.Cards.GetCard(ctx(), card.ID)
	if got2.Back != "A compiled language" {
		t.Errorf("UpdateCard back = %q, want %q", got2.Back, "A compiled language")
	}

	// Delete
	if err := s.Cards.DeleteCard(ctx(), card.ID); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}
	_, err = s.Cards.GetCard(ctx(), card.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetCard after delete = %v, want ErrNotFound", err)
	}
}

func TestCardNotFound(t *testing.T) {
	s := New()

	_, err := s.Cards.GetCard(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetCard error = %v, want ErrNotFound", err)
	}

	err = s.Cards.UpdateCard(ctx(), &domain.Card{ID: "nonexistent"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateCard error = %v, want ErrNotFound", err)
	}

	err = s.Cards.DeleteCard(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteCard error = %v, want ErrNotFound", err)
	}
}

func TestListCards_DueOnly(t *testing.T) {
	s := New()
	deck := mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	now := time.Now()
	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "due", Back: "b",
		Due: now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("setup due card: %v", err)
	}
	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "future", Back: "b",
		Due: now.Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("setup future card: %v", err)
	}

	all, _ := s.Cards.ListCards(ctx(), deck.ID, false, now)
	if len(all) != 2 {
		t.Fatalf("ListCards(all) = %d, want 2", len(all))
	}

	due, _ := s.Cards.ListCards(ctx(), deck.ID, true, now)
	if len(due) != 1 {
		t.Fatalf("ListCards(dueOnly) = %d, want 1", len(due))
	}
	if due[0].Front != "due" {
		t.Errorf("due card front = %q, want %q", due[0].Front, "due")
	}
}

func TestBulkCreateCards(t *testing.T) {
	s := New()
	deck := mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	cards := []*domain.Card{
		{DeckID: deck.ID, Front: "Q1", Back: "A1"},
		{DeckID: deck.ID, Front: "Q2", Back: "A2"},
		{DeckID: deck.ID, Front: "Q3", Back: "A3"},
	}
	if err := s.Cards.BulkCreateCards(ctx(), cards); err != nil {
		t.Fatalf("BulkCreateCards: %v", err)
	}

	all, _ := s.Cards.ListCards(ctx(), deck.ID, false, time.Now())
	if len(all) != 3 {
		t.Fatalf("ListCards after bulk = %d, want 3", len(all))
	}
	for _, c := range cards {
		if c.ID == "" {
			t.Error("BulkCreateCards did not assign ID")
		}
	}
}

// --- DeckStats tests ---

func TestDeckStats(t *testing.T) {
	s := New()
	deck := mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	now := time.Now()
	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "q1", Back: "a1",
		Due: now.Add(-time.Minute),
	}); err != nil {
		t.Fatalf("setup card: %v", err)
	}
	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "q2", Back: "a2",
		Due: now.Add(24 * time.Hour),
		SRS: domain.SRSData{State: domain.StateReview},
	}); err != nil {
		t.Fatalf("setup card: %v", err)
	}

	stats, err := s.Decks.DeckStats(ctx(), deck.ID, now)
	if err != nil {
		t.Fatalf("DeckStats: %v", err)
	}
	if stats.TotalCards != 2 {
		t.Errorf("TotalCards = %d, want 2", stats.TotalCards)
	}
	if stats.DueCards != 1 {
		t.Errorf("DueCards = %d, want 1", stats.DueCards)
	}
	if stats.NewCards != 1 {
		t.Errorf("NewCards = %d, want 1", stats.NewCards)
	}
}

// --- ReviewLog tests ---

func TestReviewLogs(t *testing.T) {
	s := New()
	now := time.Now()

	for _, log := range []*domain.ReviewLog{
		{CardID: "c1", DeckID: "d1", Rating: domain.RatingGood, ReviewedAt: now},
		{CardID: "c2", DeckID: "d2", Rating: domain.RatingEasy, ReviewedAt: now},
		{CardID: "c3", DeckID: "d1", Rating: domain.RatingAgain, ReviewedAt: now.Add(-48 * time.Hour)},
	} {
		if err := s.Reviews.CreateReviewLog(ctx(), log); err != nil {
			t.Fatalf("setup CreateReviewLog: %v", err)
		}
	}

	// All logs since yesterday
	logs, _ := s.Reviews.ListReviewLogs(ctx(), "", now.Add(-24*time.Hour))
	if len(logs) != 2 {
		t.Fatalf("ListReviewLogs(all, 24h) = %d, want 2", len(logs))
	}

	// Filtered by deck
	logs, _ = s.Reviews.ListReviewLogs(ctx(), "d1", now.Add(-72*time.Hour))
	if len(logs) != 2 {
		t.Fatalf("ListReviewLogs(d1) = %d, want 2", len(logs))
	}
}

// --- Mutation safety test ---

func TestStoreCopiesPreventsExternalMutation(t *testing.T) {
	s := New()
	deck := mustCreateDeck(t, &domain.Deck{Name: "Go"}, s.Decks)

	// Mutate the returned deck
	deck.Name = "MUTATED"

	// Fetch again — should still be "Go"
	got, _ := s.Decks.GetDeck(ctx(), deck.ID)
	if got.Name != "Go" {
		t.Errorf("store was mutated externally: got name %q", got.Name)
	}
}
