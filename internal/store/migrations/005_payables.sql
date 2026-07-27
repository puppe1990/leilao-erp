-- up
CREATE TABLE IF NOT EXISTS payables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
    due_on TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'paid', 'cancelled')),
    lot_id INTEGER REFERENCES lots(id),
    paid_at TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS payables;
