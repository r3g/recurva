# What Happens When You Commit (and Push)

This doc walks through every step that runs from the moment you type `git commit` to the moment your code is validated on GitHub. Understanding this pipeline is key to knowing what's protecting your codebase.

---

## Phase 1: Local — Before the Commit

### What you should run manually (or automate later)

Right now there are **no git hooks** installed, so nothing runs automatically before a commit. That means it's on you (or a script) to run checks before committing. Here's what the `Makefile` gives you:

```bash
make check    # runs fmt, vet, lint, test — in that order
```

That single command runs four things:

#### 1. `make fmt` → `gofumpt -w .`

**What it does:** Formats all `.go` files in the project using `gofumpt` (a stricter version of `gofmt`).

- Enforces consistent indentation, spacing, import grouping, and line breaks
- The `-w` flag writes changes back to the files (modifies them in place)
- If any files were reformatted, you'd need to `git add` them again before committing
- Config: `extra-rules: true` in `.golangci.yaml` enables additional formatting rules

**Why it matters:** Prevents style debates. Every contributor's code looks the same.

#### 2. `make vet` → `go vet ./...`

**What it does:** Runs Go's built-in static analyzer across all packages (`./...` means "this package and all sub-packages recursively").

- Catches bugs the compiler misses: suspicious function calls, unreachable code, incorrect format strings, misuse of mutexes, struct tags that won't work, etc.
- Fast — takes about 1-2 seconds
- Zero configuration needed

**Why it matters:** These are almost always real bugs, not style nits. If `go vet` complains, something is actually wrong.

#### 3. `make lint` → `golangci-lint run ./...`

**What it does:** Runs a meta-linter that combines 15 individual linters in a single pass. Configuration is in `.golangci.yaml`.

The enabled linters and what they catch:

| Linter | What it catches |
|--------|----------------|
| `errcheck` | Unchecked error return values (e.g. calling a function that returns an error and ignoring it) |
| `govet` | Same as `go vet` above (included for completeness) |
| `staticcheck` | Advanced static analysis — deprecated APIs, impossible conditions, unnecessary conversions |
| `unused` | Unused functions, variables, types, constants |
| `gosimple` | Code that could be simplified (e.g. `if x == true` → `if x`) |
| `ineffassign` | Assignments to variables that are never read afterward |
| `typecheck` | Type-checking errors (essentially what the compiler does) |
| `gofumpt` | Checks formatting matches `gofumpt` rules (without modifying files) |
| `goconst` | Strings or numbers repeated 3+ times that should be constants |
| `misspell` | Common misspellings in comments and strings |
| `gosec` | Security issues — SQL injection, hardcoded credentials, weak crypto |
| `errname` | Error type names should end in "Error", error variables should start with "Err" |
| `errorlint` | Correct use of `errors.Is()`, `errors.As()`, `%w` wrapping |
| `exhaustive` | Switch statements that don't cover all enum cases |
| `copyloopvar` | Loop variable capture bugs (copying loop vars in closures) |

**Exclusions:**
- `G115` (integer overflow) is disabled — too noisy for SRS uint32/uint64 casts
- The `imported/` and `testdata/` directories are excluded from linting

**Why it matters:** This is your most thorough local check. It catches real bugs, security issues, and code quality problems before they ever leave your machine.

#### 4. `make test` → `go test ./... -count=1`

**What it does:** Runs all tests across all packages.

- `./...` = all packages recursively
- `-count=1` = disable test caching (always run fresh, never skip "already passed")
- Currently 56+ tests covering card service, review service, SRS scheduler, domain logic, and store implementations
- Tests use an in-memory store (not SQLite), so they're fast and don't need a database

**Why it matters:** Verifies your code actually works. Tests catch regressions — changes that accidentally break existing functionality.

---

## Phase 2: The Commit Itself

### `git add <files>` — Stage Changes

Marks specific files to be included in the next commit. Git tracks changes in three areas:

```
Working Directory  →  Staging Area (Index)  →  Repository (Commits)
    (your edits)        (git add)                (git commit)
```

- `git add internal/service/card_service.go` — stage one file
- `git add -A` — stage everything (careful: could include secrets or binaries)
- Files in `.gitignore` are never staged (e.g. `*.db`, `imported/`, `recurva.db`)

### `git commit -m "message"` — Create the Commit

Records a snapshot of all staged changes with a message.

- Creates a unique SHA hash (e.g. `e4bc66a`) identifying this commit forever
- Records: who (author), when (timestamp), what (diff), why (message)
- The commit is **local only** — nothing has gone to GitHub yet
- If you have pre-commit hooks installed (you don't currently), they would run here and could block the commit

---

## Phase 3: The Push

### `git push origin main` — Send to GitHub

Uploads your local commits to the remote repository (`origin` = github.com/r3g/recurva).

- Sends only the new commits (not the entire repo)
- If someone else pushed first, you'll get a rejection and need to pull/rebase first
- This is the point of no return for making your code visible to others

---

## Phase 4: GitHub Actions CI — Automated Checks

The moment your push lands on `main` (or you open a PR targeting `main`), GitHub Actions triggers the CI pipeline defined in `.github/workflows/ci.yml`.

**Four jobs run in parallel** on fresh Ubuntu VMs:

### Job 1: `build`
```
1. Check out your code (actions/checkout@v4)
2. Install Go (version from go.mod)
3. Run: go build ./...
```
**What it proves:** Your code compiles. If you have syntax errors or missing imports, this fails.

### Job 2: `test`
```
1. Check out your code
2. Install Go
3. Run: go test ./... -count=1 -race
```
**What it proves:** All tests pass. The `-race` flag enables Go's race detector — it instruments the binary to detect concurrent access to shared data without proper synchronization. This catches bugs you'd never find in normal testing.

### Job 3: `lint`
```
1. Check out your code
2. Install Go
3. Install golangci-lint (latest)
4. Run: golangci-lint run ./...
```
**What it proves:** Same linter checks as `make lint` locally. Catches anything you forgot to check.

### Job 4: `vet`
```
1. Check out your code
2. Install Go
3. Run: go vet ./...
```
**What it proves:** Same as `make vet` locally. Redundant with the lint job (which includes govet), but provides a clear signal if only vet-level issues exist.

**If any job fails:** GitHub shows a red X on the commit. If it's a PR, the merge button shows the failure. Nothing blocks the merge on `main` pushes currently (no branch protection rules requiring CI to pass), but you'd see the failure.

---

## Phase 5: Release (Only on Tags)

When you push a **version tag** (e.g. `git tag v0.2.0 && git push --tags`), a separate workflow (`.github/workflows/release.yml`) runs:

```
1. Check out code (full history with fetch-depth: 0)
2. Install Go
3. Install GoReleaser
4. Run: goreleaser release --clean
```

**GoReleaser** (configured in `.goreleaser.yaml`):
1. Builds binaries for 4 targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
2. Creates `.tar.gz` archives for each
3. Generates a checksum file
4. Creates a GitHub Release with the archives attached
5. Auto-generates a changelog from commit messages (excluding docs/test/ci commits)

**Result:** Users can download pre-built binaries from the GitHub Releases page.

---

## The Full Timeline

```
You edit code
    ↓
make check              ← LOCAL: fmt → vet → lint → test
    ↓
git add <files>         ← Stage changes
    ↓
git commit -m "msg"     ← Create local commit
    ↓
git push origin main    ← Send to GitHub
    ↓
GitHub Actions CI       ← REMOTE: build + test + lint + vet (parallel)
    ↓                      All 4 must pass for green check
[optional]
    ↓
git tag v0.x.0          ← Tag a release
git push --tags
    ↓
GoReleaser              ← REMOTE: build binaries → create GitHub Release
```

---

## What's Missing (TODO)

- **Pre-commit hook:** A git hook that runs `make check` automatically before every commit, so you can't accidentally commit broken code. This would catch issues before they even enter the git history.
- **Pre-push hook:** Alternatively, run checks before push instead of before commit (faster iteration, same safety net).
- **Automation scripts:** Shell scripts to streamline the commit → push → verify workflow.
