-- up
CREATE TABLE IF NOT EXISTS sale_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items(id),
    unit_cost_cents_at_sale INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'main' CHECK (role IN ('main', 'accessory')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sale_lines_sale ON sale_lines(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_lines_item ON sale_lines(item_id);

-- Backfill one line per existing sale
INSERT INTO sale_lines (sale_id, item_id, unit_cost_cents_at_sale, role)
SELECT id, item_id, unit_cost_cents_at_sale, 'main'
FROM sales
WHERE NOT EXISTS (
    SELECT 1 FROM sale_lines sl WHERE sl.sale_id = sales.id
);

-- down
DROP TABLE IF EXISTS sale_lines;
