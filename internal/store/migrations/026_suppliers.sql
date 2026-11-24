-- up
CREATE TABLE IF NOT EXISTS suppliers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    whatsapp TEXT NOT NULL,
    chat_id TEXT NOT NULL UNIQUE,
    notes TEXT NOT NULL DEFAULT '',
    min_order_cents INTEGER NOT NULL DEFAULT 0 CHECK (min_order_cents >= 0),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_suppliers_created_at ON suppliers(created_at);

CREATE TABLE IF NOT EXISTS supplier_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    wa_message_id TEXT NOT NULL UNIQUE,
    from_me INTEGER NOT NULL CHECK (from_me IN (0, 1)),
    body TEXT NOT NULL DEFAULT '',
    sent_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_supplier_messages_supplier ON supplier_messages(supplier_id, sent_at, id);

-- down
DROP INDEX IF EXISTS idx_supplier_messages_supplier;
DROP TABLE IF EXISTS supplier_messages;
DROP INDEX IF EXISTS idx_suppliers_created_at;
DROP TABLE IF EXISTS suppliers;
