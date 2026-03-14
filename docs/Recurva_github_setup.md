# Recurva — GitHub Repository Setup

This document records the exact steps taken to set up the Recurva GitHub repository and make the initial commit.

## Prerequisites

- Go 1.26+ installed
- Homebrew (macOS)

## 1. Install GitHub CLI

```bash
brew install gh
```

## 2. Authenticate with GitHub

```bash
gh auth login
```

When prompted:
- **What account?** → `GitHub.com`
- **Preferred protocol?** → `HTTPS`
- **How would you like to authenticate?** → `Login with a web browser`

Follow the browser flow to authenticate as the target GitHub user (`r3g`).

Verify:

```bash
gh auth status
```

Expected output should show `Logged in to github.com account r3g`.

## 3. Update Go module path

The module was initially created with a local path. Before publishing, the module path was updated to match the GitHub repo:

```bash
# Replace in go.mod and all .go files
find . -name '*.go' -exec sed -i '' 's|github.com/y812535/recurva|github.com/r3g/recurva|g' {} +
sed -i '' 's|github.com/y812535/recurva|github.com/r3g/recurva|g' go.mod
```

Verified with:

```bash
go build ./...
go test ./... -count=1
```

## 4. Create project files

The following files were added before the initial commit:

- **`LICENSE`** — MIT license
- **`README.md`** — project description, install instructions, quick start, dev commands
- **`.gitignore`** — excludes binaries, IDE files, OS files, `.db` files, `imported/` directory, `coverage.out`

## 5. Configure git identity

Set identity scoped to this repo only (not global, to avoid affecting work repos):

```bash
git config user.name "r3g"
git config user.email "respinola@gmail.com"
```

Note: `--global` was intentionally not used.

## 6. Initialize git and create initial commit

```bash
git init
git add -A
```

Before committing, verified that `git status` showed:
- All source files staged
- `imported/` directory excluded by `.gitignore`
- No secrets or scratch data included

```bash
git commit -m "Initial commit: Recurva SRS CLI + TUI

Spaced repetition system using FSRS v4 algorithm.

- Domain layer: Card, Deck, ReviewSession, SRSData types
- Scheduler: pluggable interface with FSRS v4 adapter
- Store: SQLite (pure Go, no CGo) + in-memory for tests
- Service: deck, card, review orchestration
- TUI: Bubble Tea with menu, deck browser, review session, results
- CLI: Cobra commands (review, add, decks, cards, stats, import)
- Import: CSV and colon-delimited vocab formats
- Tests: 49 tests across domain, scheduler, store, service layers
- Tooling: Makefile, golangci-lint, gofumpt

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

Result: 42 files, 4117 insertions.

## 7. Create GitHub repo and push

Single command to create the repo, set the remote, and push:

```bash
gh repo create r3g/recurva \
  --public \
  --source=. \
  --remote=origin \
  --description "Aiming to bend the forgetting curve. A spaced repetition system built in Go." \
  --push
```

This:
- Created a public repo at `https://github.com/r3g/recurva`
- Added `origin` remote pointing to it
- Pushed the `main` branch
- Set `main` to track `origin/main`

## 8. Verify

```bash
# Remote is correct
git remote -v
# origin  https://github.com/r3g/recurva.git (fetch)
# origin  https://github.com/r3g/recurva.git (push)

# Branch tracks remote
git status
# On branch main
# Your branch is up to date with 'origin/main'.
```

Repo live at: https://github.com/r3g/recurva
