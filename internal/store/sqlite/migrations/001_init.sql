CREATE TABLE IF NOT EXISTS decks (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cards (
    id             TEXT PRIMARY KEY,
    deck_id        TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    front          TEXT NOT NULL,
    back           TEXT NOT NULL,
    notes          TEXT NOT NULL DEFAULT '',
    tags           TEXT NOT NULL DEFAULT '[]',
    due            DATETIME NOT NULL,
    stability      REAL NOT NULL DEFAULT 0,
    difficulty     REAL NOT NULL DEFAULT 0,
    elapsed_days   INTEGER NOT NULL DEFAULT 0,
    scheduled_days INTEGER NOT NULL DEFAULT 0,
    reps           INTEGER NOT NULL DEFAULT 0,
    lapses         INTEGER NOT NULL DEFAULT 0,
    state          INTEGER NOT NULL DEFAULT 0,
    last_review    DATETIME,
    created_at     DATETIME NOT NULL,
    updated_at     DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cards_deck_due ON cards(deck_id, due);
