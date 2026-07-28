-- up
ALTER TABLE products ADD COLUMN description TEXT;
ALTER TABLE products ADD COLUMN listing_text TEXT;

-- down
-- SQLite: columns left in place on rollback of app logic
