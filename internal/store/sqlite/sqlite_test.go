package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/r3g/recurva/internal/domain"
)

func ctx() context.Context { return context.Background() }

func testDB(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustCreateDeckSQL(t *testing.T, ds *DeckStore, name string) *domain.Deck {
	t.Helper()
	d, err := ds.CreateDeck(ctx(), &domain.Deck{Name: name})
	if err != nil {
		t.Fatalf("setup CreateDeck(%q): %v", name, err)
	}
	return d
}

func mustCreateCardSQL(t *testing.T, cs *CardStore, card *domain.Card) *domain.Card {
	t.Helper()
	c, err := cs.CreateCard(ctx(), card)
	if err != nil {
		t.Fatalf("setup CreateCard: %v", err)
	}
	return c
}

// --- Deck tests ---

func TestDeckCRUD(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)

	// Create
	deck, err := ds.CreateDeck(ctx(), &domain.Deck{Name: "Go", Description: "Go lang"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	if deck.ID == "" {
		t.Fatal("no ID assigned")
	}

	// Get by ID
	got, err := ds.GetDeck(ctx(), deck.ID)
	if err != nil {
		t.Fatalf("GetDeck: %v", err)
	}
	if got.Name != "Go" || got.Description != "Go lang" {
		t.Errorf("GetDeck = %+v", got)
	}

	// Get by name
	got, err = ds.GetDeckByName(ctx(), "Go")
	if err != nil {
		t.Fatalf("GetDeckByName: %v", err)
	}
	if got.ID != deck.ID {
		t.Error("GetDeckByName returned wrong deck")
	}

	// List
	decks, err := ds.ListDecks(ctx())
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("ListDecks = %d, want 1", len(decks))
	}

	// Delete
	if err := ds.DeleteDeck(ctx(), deck.ID); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}
	_, err = ds.GetDeck(ctx(), deck.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("after delete: %v, want ErrNotFound", err)
	}
}

func TestDeckDuplicateName(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)

	mustCreateDeckSQL(t, ds, "Go")
	_, err := ds.CreateDeck(ctx(), &domain.Deck{Name: "Go"})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("duplicate = %v, want ErrAlreadyExists", err)
	}
}

func TestDeckNotFound(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)

	_, err := ds.GetDeck(ctx(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetDeck = %v, want ErrNotFound", err)
	}
	err = ds.DeleteDeck(ctx(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteDeck = %v, want ErrNotFound", err)
	}
}

// --- Card tests ---

func TestCardCRUD(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")

	// Create
	card, err := cs.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID,
		Front:  "What is Go?",
		Back:   "A language",
		Tags:   []string{"basics", "intro"},
	})
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}

	// Get — verify tags roundtrip
	got, err := cs.GetCard(ctx(), card.ID)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if got.Front != "What is Go?" {
		t.Errorf("Front = %q", got.Front)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "basics" {
		t.Errorf("Tags = %v, want [basics intro]", got.Tags)
	}

	// Update with SRS data
	got.SRS.Stability = 5.5
	got.SRS.State = domain.StateReview
	got.SRS.LastReview = time.Now().UTC()
	if err := cs.UpdateCard(ctx(), got); err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	got2, _ := cs.GetCard(ctx(), card.ID)
	if got2.SRS.Stability != 5.5 {
		t.Errorf("Stability = %f, want 5.5", got2.SRS.Stability)
	}
	if got2.SRS.State != domain.StateReview {
		t.Errorf("State = %d, want StateReview", got2.SRS.State)
	}
	if got2.SRS.LastReview.IsZero() {
		t.Error("LastReview should not be zero after update")
	}

	// Delete
	if err := cs.DeleteCard(ctx(), card.ID); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}
	_, err = cs.GetCard(ctx(), card.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("after delete: %v, want ErrNotFound", err)
	}
}

func TestCardNotFound(t *testing.T) {
	db := testDB(t)
	cs := NewCardStore(db)

	_, err := cs.GetCard(ctx(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetCard = %v, want ErrNotFound", err)
	}
	err = cs.UpdateCard(ctx(), &domain.Card{ID: "nope"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateCard = %v, want ErrNotFound", err)
	}
	err = cs.DeleteCard(ctx(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteCard = %v, want ErrNotFound", err)
	}
}

func TestListCards_DueFilter(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")
	now := time.Now().UTC()

	mustCreateCardSQL(t, cs, &domain.Card{
		DeckID: deck.ID, Front: "due", Back: "b",
		Due: now.Add(-time.Hour),
	})
	mustCreateCardSQL(t, cs, &domain.Card{
		DeckID: deck.ID, Front: "future", Back: "b",
		Due: now.Add(24 * time.Hour),
	})

	all, _ := cs.ListCards(ctx(), deck.ID, false, now)
	if len(all) != 2 {
		t.Fatalf("all = %d, want 2", len(all))
	}

	due, _ := cs.ListCards(ctx(), deck.ID, true, now)
	if len(due) != 1 {
		t.Fatalf("due = %d, want 1", len(due))
	}
	if due[0].Front != "due" {
		t.Errorf("due card = %q, want %q", due[0].Front, "due")
	}
}

func TestBulkCreateCards(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")

	cards := []*domain.Card{
		{DeckID: deck.ID, Front: "Q1", Back: "A1"},
		{DeckID: deck.ID, Front: "Q2", Back: "A2"},
		{DeckID: deck.ID, Front: "Q3", Back: "A3"},
	}
	if err := cs.BulkCreateCards(ctx(), cards); err != nil {
		t.Fatalf("BulkCreateCards: %v", err)
	}

	all, _ := cs.ListCards(ctx(), deck.ID, false, time.Now())
	if len(all) != 3 {
		t.Fatalf("after bulk = %d, want 3", len(all))
	}
}

func TestDeckStats_SQLite(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")
	now := time.Now().UTC()

	mustCreateCardSQL(t, cs, &domain.Card{
		DeckID: deck.ID, Front: "q1", Back: "a1",
		Due: now.Add(-time.Minute),
	})
	mustCreateCardSQL(t, cs, &domain.Card{
		DeckID: deck.ID, Front: "q2", Back: "a2",
		Due: now.Add(24 * time.Hour),
		SRS: domain.SRSData{State: domain.StateReview},
	})

	stats, err := ds.DeckStats(ctx(), deck.ID, now)
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
	if stats.DeckName != "Go" {
		t.Errorf("DeckName = %q, want %q", stats.DeckName, "Go")
	}
}

func TestDeckStats_EmptyDeck(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)

	deck := mustCreateDeckSQL(t, ds, "Empty")

	stats, err := ds.DeckStats(ctx(), deck.ID, time.Now())
	if err != nil {
		t.Fatalf("DeckStats empty: %v", err)
	}
	if stats.TotalCards != 0 || stats.DueCards != 0 || stats.NewCards != 0 {
		t.Errorf("empty deck stats = %+v, want all zeros", stats)
	}
}

// --- ReviewLog tests ---

func TestReviewLogs_SQLite(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)
	rs := NewReviewStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")
	card := mustCreateCardSQL(t, cs, &domain.Card{DeckID: deck.ID, Front: "q", Back: "a"})

	now := time.Now().UTC()

	if err := rs.CreateReviewLog(ctx(), &domain.ReviewLog{
		CardID: card.ID, DeckID: deck.ID,
		Rating: domain.RatingGood, State: domain.StateNew,
		ReviewedAt: now,
	}); err != nil {
		t.Fatalf("setup review log: %v", err)
	}
	if err := rs.CreateReviewLog(ctx(), &domain.ReviewLog{
		CardID: card.ID, DeckID: deck.ID,
		Rating: domain.RatingEasy, State: domain.StateLearning,
		ReviewedAt: now.Add(-48 * time.Hour),
	}); err != nil {
		t.Fatalf("setup review log: %v", err)
	}

	// Since yesterday — should get 1
	logs, err := rs.ListReviewLogs(ctx(), deck.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("ListReviewLogs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs = %d, want 1", len(logs))
	}
	if logs[0].Rating != domain.RatingGood {
		t.Errorf("Rating = %v, want Good", logs[0].Rating)
	}

	// Since 3 days ago — should get 2
	logs, _ = rs.ListReviewLogs(ctx(), deck.ID, now.Add(-72*time.Hour))
	if len(logs) != 2 {
		t.Fatalf("logs = %d, want 2", len(logs))
	}
}

// --- Migration tests ---

func TestMigrationsAreIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	db1, err := Open(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	db1.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	db2.Close()
}

func TestForeignKeysEnabled(t *testing.T) {
	db := testDB(t)
	cs := NewCardStore(db)

	_, err := cs.CreateCard(ctx(), &domain.Card{
		DeckID: "nonexistent-deck",
		Front:  "q", Back: "a",
	})
	if err == nil {
		t.Fatal("expected FK violation error, got nil")
	}
}

func TestCascadeDelete(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")
	mustCreateCardSQL(t, cs, &domain.Card{DeckID: deck.ID, Front: "q", Back: "a"})

	if err := ds.DeleteDeck(ctx(), deck.ID); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}

	cards, _ := cs.ListCards(ctx(), deck.ID, false, time.Now())
	if len(cards) != 0 {
		t.Fatalf("cards after deck delete = %d, want 0 (cascade)", len(cards))
	}
}

func TestNullLastReview(t *testing.T) {
	db := testDB(t)
	ds := NewDeckStore(db)
	cs := NewCardStore(db)

	deck := mustCreateDeckSQL(t, ds, "Go")
	card := mustCreateCardSQL(t, cs, &domain.Card{
		DeckID: deck.ID, Front: "q", Back: "a",
	})

	got, err := cs.GetCard(ctx(), card.ID)
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if !got.SRS.LastReview.IsZero() {
		t.Errorf("LastReview = %v, want zero", got.SRS.LastReview)
	}
}
