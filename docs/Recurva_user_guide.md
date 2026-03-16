# Recurva User Guide

## Prerequisites

Before using the TUI, you need at least one deck with cards. Create them from the command line:

```bash
# Create a deck
recurva decks new "Vocabulary"

# Add cards one at a time
recurva add --deck Vocabulary --front "abrogate" --back "to cancel or revoke formally"

# Or import in bulk
recurva import --deck Vocabulary --format csv cards.csv
recurva import --deck Vocabulary --format vocab wordlist.txt
```

## Launching the TUI

There are two ways to start a review session:

```bash
# Launch the full TUI (menu → deck selection → review)
recurva review

# Skip straight to reviewing a specific deck
recurva review Vocabulary
```

## TUI Screens

### Main Menu

The first screen you see when running `recurva review` without a deck name.

```
Recurva
Aiming to bend the forgetting curve.

▶ Review Cards
  Browse Decks
  Quit
```

**Controls:**

| Key         | Action              |
|-------------|---------------------|
| `↑` / `k`   | Move cursor up      |
| `↓` / `j`   | Move cursor down    |
| `Enter`     | Select item         |
| `q`         | Quit                |

Both "Review Cards" and "Browse Decks" take you to the deck browser.

### Deck Browser

Shows all your decks with due card counts.

```
Decks

▶ Vocabulary  due: 42 / total: 4059
  Go          due: 3 / total: 8

↑/↓ navigate • enter review • esc back • q quit
```

**Controls:**

| Key         | Action                              |
|-------------|-------------------------------------|
| `↑` / `k`   | Move cursor up                     |
| `↓` / `j`   | Move cursor down                   |
| `Enter`     | Start review session for that deck |
| `Esc`       | Back to main menu                  |
| `q`         | Quit                               |

If you have no decks, you'll see a message telling you to create one from the command line.

### Review Session

The core of Recurva. Cards are shown one at a time in a flashcard format. The review follows a **Front → Flip → Rate** cycle for each card.

#### Step 1: Read the Front

You see the question side of the card inside a bordered box, with your progress shown above.

```
[1/42]  Deck: Vocabulary
╭──────────────────────────────────╮
│                                  │
│  abrogate (v)                    │
│                                  │
╰──────────────────────────────────╯

space to flip • esc back • q quit
```

**Controls:**

| Key     | Action                         |
|---------|--------------------------------|
| `Space` | Flip the card (reveal answer)  |
| `Esc`   | Abandon session, back to menu  |
| `q`     | Quit                           |

#### Step 2: See the Answer, Rate Yourself

After flipping, the answer appears below a divider. If the card has notes, those appear too. The rating bar shows the four options with predicted next review intervals.

```
[1/42]  Deck: Vocabulary
╭──────────────────────────────────────────╮
│                                          │
│  abrogate (v)                            │
│                                          │
│  ─────────────────                       │
│                                          │
│  to cancel or revoke formally; repeal    │
│                                          │
╰──────────────────────────────────────────╯

[1] Again (<1d)  [2] Hard (1d)  [3] Good (3d)  [4] Easy (7d)
esc unflip • q quit
```

The intervals in parentheses (e.g., `3d`, `7d`, `1mo`) tell you when you'll see this card again if you pick that rating. These are calculated by the FSRS algorithm based on your card's history.

**Controls:**

| Key       | Action                                        |
|-----------|-----------------------------------------------|
| `1` or `a` | **Again** — didn't know it, review soon       |
| `2` or `h` | **Hard** — got it but struggled               |
| `3` or `g` | **Good** — recalled correctly                 |
| `4` or `e` | **Easy** — knew it instantly                  |
| `Esc`     | Unflip (go back to front without rating)      |
| `q`       | Quit                                          |

After rating, the next card appears automatically.

#### How to Rate

- **Again**: You blanked or got it wrong. The card will come back very soon.
- **Hard**: You eventually got it, but it took effort or you weren't confident. Shorter interval than Good.
- **Good**: You recalled it correctly with reasonable effort. This is the default "I know this" rating.
- **Easy**: Instant recall, no hesitation at all. Pushes the next review further out.

When in doubt, use **Good**. Only use Easy when it truly felt effortless.

### Session Results

After rating the last card, you see a summary of how the session went.

```
Session Complete!

  Total reviewed: 42

  Again: 5
  Hard:  8
  Good:  24
  Easy:  5

Time: 12m34s

enter/q to return to menu
```

**Controls:**

| Key     | Action           |
|---------|------------------|
| `Enter` | Back to menu     |
| `q`     | Back to menu     |

### Card Editor

The card editor is available when launched via the TUI's card creation flow. It provides a form with three fields.

```
Add Card  Deck: Vocabulary

Front:
> [cursor here]

Back:
>

Notes:
>

tab to switch fields • ctrl+s to save • esc back
```

**Controls:**

| Key      | Action                        |
|----------|-------------------------------|
| `Tab`    | Move to next field            |
| `Ctrl+S` | Save card                    |
| `Esc`   | Back to menu (without saving) |

After saving, the fields clear and you can add another card immediately.

## Keyboard Reference

All keys at a glance:

| Context        | Key           | Action         |
|----------------|---------------|----------------|
| Everywhere     | `q` / `Ctrl+C` | Quit          |
| Navigation     | `↑` / `k`     | Up             |
| Navigation     | `↓` / `j`     | Down           |
| Navigation     | `Enter`       | Select         |
| Navigation     | `Esc`         | Back           |
| Review (front) | `Space`       | Flip card      |
| Review (back)  | `1` / `a`     | Rate Again     |
| Review (back)  | `2` / `h`     | Rate Hard      |
| Review (back)  | `3` / `g`     | Rate Good      |
| Review (back)  | `4` / `e`     | Rate Easy      |
| Review (back)  | `Esc`         | Unflip         |
| Editor         | `Tab`         | Next field     |
| Editor         | `Ctrl+S`      | Save card      |

## TODO

- [ ] Document `go install ./cmd/recurva` — after building from source, the `recurva` binary on your PATH is a snapshot; it won't reflect new code changes until you re-run `go install`. Use `go run ./cmd/recurva` during development to always run latest source.
- [ ] Document the Card Browser screen (`b` from deck browser) — list, filter, detail, edit, add, delete flows.

## CLI Quick Reference

These commands manage your data outside the TUI:

```bash
recurva decks list               # List all decks with stats
recurva decks new "Name"         # Create a deck
recurva decks delete "Name"      # Delete a deck and its cards

recurva add -d Name --front "Q" --back "A"   # Add a card
recurva cards list "Name"        # List cards in a deck
recurva cards delete <id>        # Delete a card by ID

recurva stats                    # Review stats (last 30 days)
recurva stats Name --days 7      # Stats for one deck, last 7 days
recurva stats --json             # Output as JSON

recurva import -d Name -f csv file.csv       # Import CSV
recurva import -d Name -f vocab file.txt     # Import vocab format
```
