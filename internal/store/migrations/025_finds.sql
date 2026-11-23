-- up
CREATE TABLE IF NOT EXISTS finds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('olx', 'facebook', 'kwara')),
    buy_cents INTEGER NOT NULL CHECK (buy_cents >= 0),
    sell_cents INTEGER NOT NULL CHECK (sell_cents >= 0),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_finds_source ON finds(source);
CREATE INDEX IF NOT EXISTS idx_finds_created_at ON finds(created_at);

-- down
DROP INDEX IF EXISTS idx_finds_created_at;
DROP INDEX IF EXISTS idx_finds_source;
DROP TABLE IF EXISTS finds;
