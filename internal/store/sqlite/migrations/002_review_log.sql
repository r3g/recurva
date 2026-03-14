CREATE TABLE IF NOT EXISTS review_logs (
    id             TEXT PRIMARY KEY,
    card_id        TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    deck_id        TEXT NOT NULL,
    rating         INTEGER NOT NULL,
    state          INTEGER NOT NULL,
    scheduled_days INTEGER NOT NULL DEFAULT 0,
    elapsed_days   INTEGER NOT NULL DEFAULT 0,
    reviewed_at    DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_logs_deck_reviewed ON review_logs(deck_id, reviewed_at);
