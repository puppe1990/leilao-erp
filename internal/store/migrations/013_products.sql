-- up
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    sale_price_hint_cents INTEGER,
    kind TEXT NOT NULL DEFAULT 'principal'
        CHECK (kind IN ('principal', 'accessory')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);

-- Add product_id to items (nullable for safety during backfill)
-- SQLite: add column if not exists via simple ALTER (fails if re-run; migrations run once)
ALTER TABLE items ADD COLUMN product_id INTEGER REFERENCES products(id);

CREATE INDEX IF NOT EXISTS idx_items_product ON items(product_id);

-- Backfill one product per distinct title
INSERT INTO products (name, sale_price_hint_cents, kind)
SELECT
    i.title,
    (
        SELECT i2.sale_price_hint_cents
        FROM items i2
        WHERE i2.title = i.title AND i2.sale_price_hint_cents IS NOT NULL
        ORDER BY i2.id
        LIMIT 1
    ),
    CASE
        WHEN lower(i.title) LIKE 'cabo%' OR lower(i.title) LIKE '% cabo%' THEN 'accessory'
        ELSE 'principal'
    END
FROM items i
GROUP BY i.title
ON CONFLICT(name) DO NOTHING;

UPDATE items
SET product_id = (
    SELECT p.id FROM products p WHERE p.name = items.title LIMIT 1
)
WHERE product_id IS NULL;

-- down
-- SQLite cannot DROP COLUMN reliably on older versions; leave product_id if rolled back manually
DROP INDEX IF EXISTS idx_items_product;
DROP INDEX IF EXISTS idx_products_name;
DROP TABLE IF EXISTS products;
