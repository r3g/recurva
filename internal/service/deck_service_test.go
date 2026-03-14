package service

import (
	"context"
	"errors"
	"testing"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store/memory"
)

func ctx() context.Context { return context.Background() }

func TestDeckService_CreateAndList(t *testing.T) {
	svc := NewDeckService(*memory.New())

	deck, err := svc.CreateDeck(ctx(), "Go", "Go programming")
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	if deck.Name != "Go" {
		t.Errorf("name = %q, want %q", deck.Name, "Go")
	}

	decks, err := svc.ListDecks(ctx())
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("ListDecks = %d, want 1", len(decks))
	}
}

func TestDeckService_CreateEmptyName(t *testing.T) {
	svc := NewDeckService(*memory.New())

	_, err := svc.CreateDeck(ctx(), "", "")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("CreateDeck('') = %v, want ErrInvalidInput", err)
	}
}

func TestDeckService_DeleteByName(t *testing.T) {
	svc := NewDeckService(*memory.New())
	if _, err := svc.CreateDeck(ctx(), "Go", ""); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := svc.DeleteDeck(ctx(), "Go"); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}

	decks, _ := svc.ListDecks(ctx())
	if len(decks) != 0 {
		t.Fatalf("after delete: %d decks, want 0", len(decks))
	}
}

func TestDeckService_DeleteNotFound(t *testing.T) {
	svc := NewDeckService(*memory.New())

	err := svc.DeleteDeck(ctx(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteDeck = %v, want ErrNotFound", err)
	}
}

func TestDeckService_GetByName(t *testing.T) {
	svc := NewDeckService(*memory.New())
	if _, err := svc.CreateDeck(ctx(), "Go", "desc"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := svc.GetDeckByName(ctx(), "Go")
	if err != nil {
		t.Fatalf("GetDeckByName: %v", err)
	}
	if got.Description != "desc" {
		t.Errorf("Description = %q, want %q", got.Description, "desc")
	}
}

func TestDeckService_AllDeckStats(t *testing.T) {
	s := memory.New()
	svc := NewDeckService(*s)

	deck, err := svc.CreateDeck(ctx(), "Go", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := s.Cards.CreateCard(ctx(), &domain.Card{
		DeckID: deck.ID, Front: "q", Back: "a",
	}); err != nil {
		t.Fatalf("setup card: %v", err)
	}

	stats, err := svc.AllDeckStats(ctx())
	if err != nil {
		t.Fatalf("AllDeckStats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("stats = %d, want 1", len(stats))
	}
	if stats[0].TotalCards != 1 {
		t.Errorf("TotalCards = %d, want 1", stats[0].TotalCards)
	}
}
