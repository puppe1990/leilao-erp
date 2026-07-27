-- up
CREATE TABLE IF NOT EXISTS receivables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
    due_on TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'received', 'cancelled')),
    sale_id INTEGER REFERENCES sales(id),
    received_at TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS receivables;
