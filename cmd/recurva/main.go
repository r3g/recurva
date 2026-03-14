package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r3g/recurva/internal/config"
	"github.com/r3g/recurva/internal/scheduler/fsrs"
	"github.com/r3g/recurva/internal/service"
	"github.com/r3g/recurva/internal/store"
	sqlitestore "github.com/r3g/recurva/internal/store/sqlite"
	"github.com/r3g/recurva/internal/tui"
	"github.com/r3g/recurva/internal/tui/review"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "recurva",
		Short: "Recurva — Aiming to bend the forgetting curve",
	}
	root.AddCommand(
		reviewCmd(),
		addCmd(),
		decksCmd(),
		cardsCmd(),
		statsCmd(),
		importCmd(),
	)
	return root
}

func loadServices() (tui.Services, func(), error) {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return tui.Services{}, nil, fmt.Errorf("config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return tui.Services{}, nil, err
	}

	db, err := sqlitestore.Open(cfg.DBPath)
	if err != nil {
		return tui.Services{}, nil, fmt.Errorf("open db: %w", err)
	}

	s := store.Store{
		Cards:   sqlitestore.NewCardStore(db),
		Decks:   sqlitestore.NewDeckStore(db),
		Reviews: sqlitestore.NewReviewStore(db),
	}

	sched := fsrs.New(cfg.ToFSRSParams())

	svc := tui.Services{
		Decks:   service.NewDeckService(s),
		Cards:   service.NewCardService(s),
		Reviews: service.NewReviewService(s, sched),
	}

	return svc, func() { db.Close() }, nil
}

func reviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "review [deck]",
		Short: "Start a review session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			if len(args) == 1 {
				// Direct deck review
				m, _ := review.New(svc.Reviews, args[0])
				p := tea.NewProgram(m, tea.WithAltScreen())
				_, err := p.Run()
				return err
			}

			// Launch TUI app
			app := tui.NewApp(svc)
			p := tea.NewProgram(app, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
}

func addCmd() *cobra.Command {
	var deckName, front, back, notes string
	var tags []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a card",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			if front != "" && back != "" && deckName != "" {
				card, err := svc.Cards.AddCard(context.Background(), deckName, front, back, notes, tags)
				if err != nil {
					return err
				}
				fmt.Printf("Card added: %s\n", card.ID)
				return nil
			}

			return fmt.Errorf("use --deck, --front, and --back flags to add a card non-interactively")
		},
	}
	cmd.Flags().StringVarP(&deckName, "deck", "d", "", "Deck name")
	cmd.Flags().StringVar(&front, "front", "", "Card front (question)")
	cmd.Flags().StringVar(&back, "back", "", "Card back (answer)")
	cmd.Flags().StringVar(&notes, "notes", "", "Card notes")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags (comma-separated)")
	return cmd
}

func decksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decks",
		Short: "Manage decks",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all decks",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			stats, err := svc.Decks.AllDeckStats(context.Background())
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTOTAL\tDUE\tNEW")
			for _, s := range stats {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", s.DeckName, s.TotalCards, s.DueCards, s.NewCards)
			}
			return w.Flush()
		},
	}

	newCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new deck",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			desc, _ := cmd.Flags().GetString("description")
			deck, err := svc.Decks.CreateDeck(context.Background(), args[0], desc)
			if err != nil {
				return err
			}
			fmt.Printf("Deck created: %s (id: %s)\n", deck.Name, deck.ID)
			return nil
		},
	}
	newCmd.Flags().StringP("description", "d", "", "Deck description")

	deleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a deck",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			return svc.Decks.DeleteDeck(context.Background(), args[0])
		},
	}

	cmd.AddCommand(listCmd, newCmd, deleteCmd)
	return cmd
}

func cardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cards",
		Short: "Manage cards",
	}

	listCmd := &cobra.Command{
		Use:   "list <deck>",
		Short: "List cards in a deck",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			cards, err := svc.Cards.ListCards(context.Background(), args[0])
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFRONT\tSTATE\tDUE")
			for _, c := range cards {
				stateStr := []string{"New", "Learning", "Review", "Relearning"}[c.SRS.State]
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					c.ID[:8]+"...",
					truncate(c.Front, 40),
					stateStr,
					c.Due.Format("2006-01-02"),
				)
			}
			return w.Flush()
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			return svc.Cards.DeleteCard(context.Background(), args[0])
		},
	}

	cmd.AddCommand(listCmd, deleteCmd)
	return cmd
}

func statsCmd() *cobra.Command {
	var days int
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "stats [deck]",
		Short: "Show review statistics",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			deckID := ""
			if len(args) == 1 {
				deck, err := svc.Decks.GetDeckByName(context.Background(), args[0])
				if err != nil {
					return err
				}
				deckID = deck.ID
			}

			logs, err := svc.Reviews.ReviewStats(context.Background(), deckID, days)
			if err != nil {
				return err
			}

			if asJSON {
				return json.NewEncoder(os.Stdout).Encode(logs)
			}

			counts := map[string]int{"Again": 0, "Hard": 0, "Good": 0, "Easy": 0}
			for _, l := range logs {
				counts[l.Rating.String()]++
			}

			fmt.Printf("Reviews in last %d days: %d\n", days, len(logs))
			fmt.Printf("  Again: %d\n  Hard:  %d\n  Good:  %d\n  Easy:  %d\n",
				counts["Again"], counts["Hard"], counts["Good"], counts["Easy"])
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 30, "Number of days to show stats for")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func importCmd() *cobra.Command {
	var deckName string
	var format string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import cards from a file",
		Long:  "Import cards from a file. Supported formats: csv (default), vocab (colon-delimited word:pos:definition:flag:id:timestamp)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := loadServices()
			if err != nil {
				return err
			}
			defer cleanup()

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			var n int
			switch format {
			case "csv":
				n, err = svc.Cards.ImportCSV(context.Background(), deckName, f)
			case "vocab":
				n, err = svc.Cards.ImportVocab(context.Background(), deckName, f)
			default:
				return fmt.Errorf("unsupported format %q (use csv or vocab)", format)
			}
			if err != nil {
				return err
			}
			fmt.Printf("Imported %d cards into deck %q\n", n, deckName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&deckName, "deck", "d", "", "Target deck name (required)")
	cmd.Flags().StringVarP(&format, "format", "f", "csv", "Import format: csv, vocab")
	_ = cmd.MarkFlagRequired("deck")
	return cmd
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

// suppress unused import
var _ = strings.TrimSpace
