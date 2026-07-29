-- up
ALTER TABLE products ADD COLUMN slug TEXT NOT NULL DEFAULT '';

-- Unique among non-empty slugs (backfill fills empties on boot).
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug ON products(slug) WHERE slug != '';

-- down
-- SQLite: column/index left in place
