-- up
ALTER TABLE products ADD COLUMN olx_published INTEGER NOT NULL DEFAULT 0;

-- down
-- SQLite: column left in place
