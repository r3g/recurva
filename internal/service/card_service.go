package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store"
)

type CardService struct {
	store store.Store
}

func NewCardService(s store.Store) *CardService {
	return &CardService{store: s}
}

func (s *CardService) AddCard(ctx context.Context, deckName, front, back, notes string, tags []string) (*domain.Card, error) {
	if front == "" || back == "" {
		return nil, domain.ErrInvalidInput
	}
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return nil, fmt.Errorf("deck %q: %w", deckName, err)
	}
	card := &domain.Card{
		DeckID: deck.ID,
		Front:  front,
		Back:   back,
		Notes:  notes,
		Tags:   tags,
	}
	return s.store.Cards.CreateCard(ctx, card)
}

func (s *CardService) ListCards(ctx context.Context, deckName string) ([]*domain.Card, error) {
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return nil, err
	}
	return s.store.Cards.ListCards(ctx, deck.ID, false, timeNow())
}

func (s *CardService) DeleteCard(ctx context.Context, id string) error {
	return s.store.Cards.DeleteCard(ctx, id)
}

func (s *CardService) ImportCSV(ctx context.Context, deckName string, r io.Reader) (int, error) {
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return 0, fmt.Errorf("deck %q: %w", deckName, err)
	}

	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1

	var cards []*domain.Card
	lineNum := 0
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("csv read: %w", err)
		}
		lineNum++
		if lineNum == 1 && strings.ToLower(record[0]) == "front" {
			continue // skip header
		}
		if len(record) < 2 {
			continue
		}
		card := &domain.Card{
			DeckID: deck.ID,
			Front:  strings.TrimSpace(record[0]),
			Back:   strings.TrimSpace(record[1]),
		}
		if len(record) >= 3 {
			card.Notes = strings.TrimSpace(record[2])
		}
		cards = append(cards, card)
	}

	if err := s.store.Cards.BulkCreateCards(ctx, cards); err != nil {
		return 0, err
	}
	return len(cards), nil
}

// ImportVocab imports cards from a colon-delimited vocabulary file.
// Expected format per line: word:pos:definition:flag:id:timestamp
// Front becomes "word (pos)", Back becomes the definition.
func (s *CardService) ImportVocab(ctx context.Context, deckName string, r io.Reader) (int, error) {
	deck, err := s.store.Decks.GetDeckByName(ctx, deckName)
	if err != nil {
		return 0, fmt.Errorf("deck %q: %w", deckName, err)
	}

	scanner := bufio.NewScanner(r)
	var cards []*domain.Card
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		card, err := parseVocabLine(line, deck.ID)
		if err != nil {
			continue // skip malformed lines
		}
		cards = append(cards, card)
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read vocab: %w", err)
	}

	if err := s.store.Cards.BulkCreateCards(ctx, cards); err != nil {
		return 0, err
	}
	return len(cards), nil
}

func parseVocabLine(line, deckID string) (*domain.Card, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("expected at least 6 colon-separated fields, got %d", len(parts))
	}

	word := strings.TrimSpace(parts[0])
	pos := strings.TrimSpace(parts[1])
	// Definition may contain colons; last 3 fields are flag, id, timestamp.
	definition := strings.TrimSpace(strings.Join(parts[2:len(parts)-3], ":"))

	if word == "" || definition == "" {
		return nil, fmt.Errorf("empty word or definition")
	}

	front := fmt.Sprintf("%s (%s)", word, pos)
	return &domain.Card{
		DeckID: deckID,
		Front:  front,
		Back:   definition,
		Tags:   []string{pos},
	}, nil
}
