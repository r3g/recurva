// Command rdef looks up a word in the local Recurva vocabulary deck and
// prints a one-line "term (pos) — definition" result.
//
// Matching order: exact (case-insensitive) → prefix → Levenshtein fuzzy.
// Read-only: opens the sqlite DB in mode=ro so it can run alongside an
// active Recurva session without contention.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/r3g/recurva/internal/config"

	_ "modernc.org/sqlite"
)

const usage = `rdef — look up a word in your Recurva vocabulary deck

Usage:
  rdef <term>

Examples:
  rdef acrimonious
  rdef "sine qua non"
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "rdef:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	term := strings.TrimSpace(args[0])
	if term == "" {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if _, err := os.Stat(cfg.DBPath); err != nil {
		return fmt.Errorf("Recurva DB not found at %s — run `recurva import` first", cfg.DBPath)
	}

	db, err := sql.Open("sqlite", "file:"+cfg.DBPath+"?mode=ro&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	hits, err := lookup(db, term)
	if err != nil {
		return err
	}
	if len(hits) == 0 {
		return fmt.Errorf("no match for %q", term)
	}
	for _, h := range hits {
		fmt.Printf("%s — %s\n", h.front, h.back)
	}
	return nil
}

type hit struct {
	front string
	back  string
}

// lookup runs the exact → prefix → fuzzy cascade. Returns an empty slice on
// total miss (never nil-with-error for "not found" — that's the caller's job
// to interpret).
func lookup(db *sql.DB, term string) ([]hit, error) {
	escTerm := escapeLike(term)

	hits, err := queryHits(db, `
		SELECT front, back FROM cards
		WHERE LOWER(front) = LOWER(?)
		   OR LOWER(front) LIKE LOWER(?) || ' (%' ESCAPE '\'
		ORDER BY front`,
		term, escTerm)
	if err != nil {
		return nil, err
	}
	if len(hits) > 0 {
		return hits, nil
	}

	hits, err = queryHits(db, `
		SELECT front, back FROM cards
		WHERE LOWER(front) LIKE LOWER(?) || '%' ESCAPE '\'
		ORDER BY LENGTH(front), front
		LIMIT 5`, escTerm)
	if err != nil {
		return nil, err
	}
	if len(hits) > 0 {
		return hits, nil
	}

	return fuzzy(db, term)
}

func queryHits(db *sql.DB, query string, args ...any) ([]hit, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []hit
	for rows.Next() {
		var h hit
		if err := rows.Scan(&h.front, &h.back); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// fuzzy scans every card's front, strips the " (pos)" suffix for comparison,
// and returns all cards tied for minimum Levenshtein distance to term — but
// only if the distance is within threshold.
func fuzzy(db *sql.DB, term string) ([]hit, error) {
	rows, err := db.Query(`SELECT front, back FROM cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lcTerm := strings.ToLower(term)
	threshold := 2
	if len([]rune(term)) > 6 {
		threshold = 3
	}

	best := -1
	var bestHits []hit
	for rows.Next() {
		var h hit
		if err := rows.Scan(&h.front, &h.back); err != nil {
			return nil, err
		}
		word := strings.ToLower(stripPOS(h.front))
		d := levenshtein(lcTerm, word)
		if d > threshold {
			continue
		}
		switch {
		case best == -1 || d < best:
			best = d
			bestHits = []hit{h}
		case d == best:
			bestHits = append(bestHits, h)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if best == -1 {
		return nil, nil
	}
	return bestHits, nil
}

// stripPOS trims a trailing " (pos)" suffix from a card front so that
// "acrimonious (adj)" becomes "acrimonious" for comparison purposes. Fronts
// without a parenthesized suffix are returned unchanged.
func stripPOS(front string) string {
	if i := strings.LastIndex(front, " ("); i >= 0 && strings.HasSuffix(front, ")") {
		return front[:i]
	}
	return front
}

// escapeLike escapes SQL LIKE wildcards so user-supplied text is matched
// literally. The ESCAPE clause in the caller must use backslash.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// levenshtein computes the edit distance between a and b with the classic
// two-row dynamic programming approach. O(len(a)*len(b)) time, O(len(b)) space.
func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}
