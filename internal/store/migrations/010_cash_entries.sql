-- up
CREATE TABLE IF NOT EXISTS cash_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES cash_accounts(id),
    direction TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
    occurred_at TEXT NOT NULL,
    category TEXT NOT NULL,
    memo TEXT,
    sale_id INTEGER REFERENCES sales(id),
    payable_id INTEGER REFERENCES payables(id),
    receivable_id INTEGER REFERENCES receivables(id),
    lot_id INTEGER REFERENCES lots(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS cash_entries;
