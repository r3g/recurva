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

## Development

```bash
make build    # compile
make test     # run tests
make lint     # golangci-lint
make fmt      # gofumpt
make check    # all of the above
```

## License

MIT
