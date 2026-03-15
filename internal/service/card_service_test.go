package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store/memory"
)

func setupCardTest(t *testing.T) (*CardService, *DeckService) {
	t.Helper()
	s := memory.New()
	deckSvc := NewDeckService(*s)
	cardSvc := NewCardService(*s)
	if _, err := deckSvc.CreateDeck(ctx(), "Go", ""); err != nil {
		t.Fatalf("setup deck: %v", err)
	}
	return cardSvc, deckSvc
}

func TestCardService_AddCard(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	card, err := cardSvc.AddCard(ctx(), "Go", "What is Go?", "A language", "", nil)
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if card.Front != "What is Go?" {
		t.Errorf("Front = %q", card.Front)
	}
}

func TestCardService_AddCard_EmptyFields(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	_, err := cardSvc.AddCard(ctx(), "Go", "", "answer", "", nil)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty front = %v, want ErrInvalidInput", err)
	}

	_, err = cardSvc.AddCard(ctx(), "Go", "question", "", "", nil)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty back = %v, want ErrInvalidInput", err)
	}
}

func TestCardService_AddCard_DeckNotFound(t *testing.T) {
	cardSvc := NewCardService(*memory.New())

	_, err := cardSvc.AddCard(ctx(), "Nonexistent", "q", "a", "", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want wrapped ErrNotFound", err)
	}
}

func TestCardService_ListCards(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	if _, err := cardSvc.AddCard(ctx(), "Go", "Q1", "A1", "", nil); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := cardSvc.AddCard(ctx(), "Go", "Q2", "A2", "", nil); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cards, err := cardSvc.ListCards(ctx(), "Go")
	if err != nil {
		t.Fatalf("ListCards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("ListCards = %d, want 2", len(cards))
	}
}

func TestCardService_DeleteCard(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	card, err := cardSvc.AddCard(ctx(), "Go", "Q", "A", "", nil)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := cardSvc.DeleteCard(ctx(), card.ID); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}

	cards, _ := cardSvc.ListCards(ctx(), "Go")
	if len(cards) != 0 {
		t.Fatalf("after delete: %d, want 0", len(cards))
	}
}

func TestCardService_ImportCSV(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	csv := "front,back,notes\nWhat is Go?,A language,Created by Google\nWhat is a goroutine?,A lightweight thread,\n"
	n, err := cardSvc.ImportCSV(ctx(), "Go", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportCSV: %v", err)
	}
	if n != 2 {
		t.Fatalf("imported = %d, want 2", n)
	}

	cards, _ := cardSvc.ListCards(ctx(), "Go")
	if len(cards) != 2 {
		t.Fatalf("ListCards = %d, want 2", len(cards))
	}
}

func TestCardService_ImportCSV_NoHeader(t *testing.T) {
	cardSvc, _ := setupCardTest(t)

	csv := "What is Go?,A language\nWhat is a goroutine?,A lightweight thread\n"
	n, err := cardSvc.ImportCSV(ctx(), "Go", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportCSV: %v", err)
	}
	if n != 2 {
		t.Fatalf("imported = %d, want 2", n)
	}
}

func TestCardService_ImportCSV_DeckNotFound(t *testing.T) {
	cardSvc := NewCardService(*memory.New())

	_, err := cardSvc.ImportCSV(ctx(), "Nonexistent", strings.NewReader("q,a\n"))
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestCardService_ImportVocab(t *testing.T) {
	s := memory.New()
	deckSvc := NewDeckService(*s)
	cardSvc := NewCardService(*s)
	if _, err := deckSvc.CreateDeck(ctx(), "Vocab", ""); err != nil {
		t.Fatalf("setup: %v", err)
	}

	input := strings.Join([]string{
		"abase:v:to cause to feel shame; hurt the pride of; (a):gmat|gre|sat",
		"abash:v:to destroy the self-possession or self-confidence of; embarrass:",
		"abbey:n:the group of buildings which collectively form the dwelling-place of a society of monks or nuns; nunnery; priory:",
	}, "\n")

	n, err := cardSvc.ImportVocab(ctx(), "Vocab", strings.NewReader(input))
	if err != nil {
		t.Fatalf("ImportVocab: %v", err)
	}
	if n != 3 {
		t.Fatalf("imported = %d, want 3", n)
	}

	cards, _ := cardSvc.ListCards(ctx(), "Vocab")
	if len(cards) != 3 {
		t.Fatalf("ListCards = %d, want 3", len(cards))
	}

	// Verify card format
	var found bool
	for _, c := range cards {
		if c.Front == "abase (v)" {
			found = true
			if !strings.Contains(c.Back, "to cause to feel shame") {
				t.Errorf("abase back = %q, expected definition", c.Back)
			}
			wantTags := []string{"v", "gmat", "gre", "sat"}
			if len(c.Tags) != len(wantTags) {
				t.Errorf("abase tags = %v, want %v", c.Tags, wantTags)
			} else {
				for i, tag := range wantTags {
					if c.Tags[i] != tag {
						t.Errorf("abase tags[%d] = %q, want %q", i, c.Tags[i], tag)
					}
				}
			}
		}
	}
	if !found {
		t.Error("card 'abase (v)' not found in imported cards")
	}
}

func TestCardService_ImportVocab_SkipsMalformed(t *testing.T) {
	s := memory.New()
	deckSvc := NewDeckService(*s)
	cardSvc := NewCardService(*s)
	if _, err := deckSvc.CreateDeck(ctx(), "Vocab", ""); err != nil {
		t.Fatalf("setup: %v", err)
	}

	input := "good:v:definition:gre\nbadline\n\n"
	n, err := cardSvc.ImportVocab(ctx(), "Vocab", strings.NewReader(input))
	if err != nil {
		t.Fatalf("ImportVocab: %v", err)
	}
	if n != 1 {
		t.Fatalf("imported = %d, want 1 (skip malformed)", n)
	}
}

func TestCardService_ImportVocab_DeckNotFound(t *testing.T) {
	cardSvc := NewCardService(*memory.New())

	_, err := cardSvc.ImportVocab(ctx(), "Nonexistent", strings.NewReader("a:v:b:gre\n"))
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestParseVocabLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantFront string
		wantBack  string
		wantTags  []string
		wantErr   bool
	}{
		{
			name:      "standard with flags",
			line:      "abase:v:to cause to feel shame:gmat|gre|sat",
			wantFront: "abase (v)",
			wantBack:  "to cause to feel shame",
			wantTags:  []string{"v", "gmat", "gre", "sat"},
		},
		{
			name:      "empty flag",
			line:      "abbey:n:a dwelling place:",
			wantFront: "abbey (n)",
			wantBack:  "a dwelling place",
			wantTags:  []string{"n"},
		},
		{
			name:      "single flag",
			line:      "abet:v:to encourage:gre",
			wantFront: "abet (v)",
			wantBack:  "to encourage",
			wantTags:  []string{"v", "gre"},
		},
		{
			name:      "definition with colon",
			line:      "word:adj:means this: or that:sat",
			wantFront: "word (adj)",
			wantBack:  "means this: or that",
			wantTags:  []string{"adj", "sat"},
		},
		{
			name:    "too few fields",
			line:    "word:v:def",
			wantErr: true,
		},
		{
			name:    "empty word",
			line:    ":v:definition:gre",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := parseVocabLine(tt.line, "deck-1")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Front != tt.wantFront {
				t.Errorf("Front = %q, want %q", card.Front, tt.wantFront)
			}
			if card.Back != tt.wantBack {
				t.Errorf("Back = %q, want %q", card.Back, tt.wantBack)
			}
			if tt.wantTags != nil {
				if len(card.Tags) != len(tt.wantTags) {
					t.Errorf("Tags = %v, want %v", card.Tags, tt.wantTags)
				} else {
					for i, tag := range tt.wantTags {
						if card.Tags[i] != tag {
							t.Errorf("Tags[%d] = %q, want %q", i, card.Tags[i], tag)
						}
					}
				}
			}
			if card.DeckID != "deck-1" {
				t.Errorf("DeckID = %q, want %q", card.DeckID, "deck-1")
			}
		})
	}
}
