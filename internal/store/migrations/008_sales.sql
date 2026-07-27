-- up
CREATE TABLE IF NOT EXISTS sales (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id INTEGER NOT NULL REFERENCES items(id),
    sold_at TEXT NOT NULL,
    channel TEXT NOT NULL CHECK (channel IN ('direct', 'mercadolivre', 'shopee', 'olx', 'other')),
    gross_cents INTEGER NOT NULL,
    fee_cents INTEGER NOT NULL DEFAULT 0,
    shipping_cents INTEGER NOT NULL DEFAULT 0,
    net_cents INTEGER NOT NULL,
    payment_status TEXT NOT NULL CHECK (payment_status IN ('received', 'pending', 'cancelled')),
    unit_cost_cents_at_sale INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS sales;
