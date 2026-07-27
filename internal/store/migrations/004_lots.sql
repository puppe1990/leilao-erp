-- up
CREATE TABLE IF NOT EXISTS lots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    auction_source TEXT,
    purchased_at TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'partial', 'sold', 'closed')),
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS lots;
