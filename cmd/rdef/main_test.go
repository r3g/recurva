package main

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"acrimonious", "acrimonius", 1},
		{"acrimonious", "acrimonious", 0},
		{"abrogate", "abrogation", 3},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestStripPOS(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"acrimonious (adj)", "acrimonious"},
		{"run (v)", "run"},
		{"sine qua non (phr)", "sine qua non"},
		{"plain", "plain"},
		{"weird (", "weird ("},
	}
	for _, c := range cases {
		if got := stripPOS(c.in); got != c.want {
			t.Errorf("stripPOS(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeLike(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"abc", "abc"},
		{"50%", `50\%`},
		{"a_b", `a\_b`},
		{`x\y`, `x\\y`},
	}
	for _, c := range cases {
		if got := escapeLike(c.in); got != c.want {
			t.Errorf("escapeLike(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

// TestLookup exercises the three match tiers against an in-memory sqlite
// database that mimics the Recurva cards schema.
func TestLookup(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE cards (id TEXT, front TEXT, back TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	seed := []struct{ front, back string }{
		{"acrimonious (adj)", "bitter, caustic, or sharply critical in tone"},
		{"abrogate (v)", "to repeal or do away with formally"},
		{"abrogation (n)", "the act of abrogating"},
		{"run (n)", "an act of running"},
		{"run (v)", "to move swiftly on foot"},
		{"zygote (n)", "a fertilized egg cell"},
	}
	for _, s := range seed {
		if _, err := db.Exec(`INSERT INTO cards(id, front, back) VALUES(?,?,?)`, s.front, s.front, s.back); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("exact single POS", func(t *testing.T) {
		got, err := lookup(db, "acrimonious")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].front != "acrimonious (adj)" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("exact multi POS", func(t *testing.T) {
		got, err := lookup(db, "run")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("want 2 POS variants, got %+v", got)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got, err := lookup(db, "ACRIMONIOUS")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("prefix fallback", func(t *testing.T) {
		got, err := lookup(db, "abrog")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Fatal("expected prefix matches")
		}
		// Shortest-first ordering: abrogate before abrogation.
		if got[0].front != "abrogate (v)" {
			t.Fatalf("want abrogate first, got %+v", got)
		}
	})

	t.Run("fuzzy fallback", func(t *testing.T) {
		got, err := lookup(db, "acrimoniuos") // transposed typo
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].front != "acrimonious (adj)" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("total miss", func(t *testing.T) {
		got, err := lookup(db, "xyzzyplover")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("expected no hits, got %+v", got)
		}
	})
}
