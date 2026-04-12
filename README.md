# Recurva

Aiming to bend the forgetting curve.

Recurva is a spaced repetition system (SRS) built in Go. It uses the [FSRS v4](https://github.com/open-spaced-repetition/go-fsrs) algorithm to schedule reviews and help you retain what you learn.

## Install

```bash
go install github.com/r3g/recurva/cmd/recurva@latest
```

## Quick Start

```bash
# Create a deck
recurva decks new "Go"

# Add cards
recurva add --deck Go --front "What is a goroutine?" --back "A lightweight thread managed by the Go runtime"

# Import from CSV
recurva import --deck Go --format csv cards.csv

# Import vocabulary (colon-delimited format)
recurva import --deck Vocab --format vocab wordlist.txt

# Review
recurva review Go

# Stats
recurva stats Go
```

## `rdef` — instant word lookup

`rdef` is a companion binary that looks up a word in your local Recurva
vocabulary database and prints `term (pos) — definition` to stdout. It runs
read-only against the same SQLite file your main Recurva install uses, so it
works alongside an active review session.

```bash
rdef acrimonious
# acrimonious (adj) — (typically of speech or discussion) angry and bitter.
```

Match order:

1. **Exact** case-insensitive (matches `word` or `word (pos)`)
2. **Prefix** — top 5 shortest matches starting with your input
3. **Fuzzy** — closest Levenshtein neighbour (threshold 2 for ≤6-char inputs, 3 otherwise)

Words with multiple parts of speech print one line each. Total miss prints
`rdef: no match for "<term>"` to stderr and exits `1`.

### Installing `rdef` from a fresh clone

These steps assume you are setting up on a brand new machine with nothing but
macOS (or Linux) and a terminal.

**1. Install Go 1.26 or later.**

```bash
# macOS via Homebrew
brew install go

# or download from https://go.dev/dl/
go version   # confirm 1.26+
```

**2. Clone the repo.**

```bash
git clone https://github.com/r3g/recurva.git
cd recurva
```

**3. Build and install both binaries.**

```bash
make install-all
```

This runs `go install ./cmd/recurva ./cmd/rdef`, which compiles both binaries
and drops them into `$(go env GOPATH)/bin` (default `~/go/bin`).

If you only want `rdef`:

```bash
make install-rdef
# or, equivalently:
go install github.com/r3g/recurva/cmd/rdef
```

**4. Put `~/go/bin` on your `PATH`.**

Check first:

```bash
echo $PATH | tr ':' '\n' | grep -q "$(go env GOPATH)/bin" && echo "already on PATH" || echo "needs adding"
```

If it needs adding, append one line to your shell rc file. For **zsh** (macOS default):

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

For **bash**:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

Verify:

```bash
which rdef
# /Users/you/go/bin/rdef
```

**5. Populate the vocabulary database.**

`rdef` reads the same SQLite database Recurva uses. By default it lives at
`~/.local/share/recurva/recurva.db`. On a fresh clone this file does not
exist yet — you need to import vocab first:

```bash
recurva decks new Vocab
recurva import --deck Vocab --format vocab imported/new_jawnz.txt
```

(Any colon-delimited `word:pos:definition:flags` file works — see
`imported/` for examples.)

**6. Look up a word from anywhere.**

```bash
cd ~/anywhere
rdef perspicacious
# perspicacious (adj) — having a ready insight into and understanding of things
```

### Uninstalling `rdef`

```bash
rm "$(go env GOPATH)/bin/rdef"
```

## Development

```bash
make build         # compile everything
make test          # run tests
make lint          # golangci-lint
make fmt           # gofumpt
make check         # all of the above
make install       # install cmd/recurva
make install-rdef  # install cmd/rdef
make install-all   # install both
```

## License

MIT
