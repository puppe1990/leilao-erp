-- up
ALTER TABLE sales ADD COLUMN client_id INTEGER REFERENCES clients(id);
CREATE INDEX IF NOT EXISTS idx_sales_client_id ON sales(client_id);

-- down
DROP INDEX IF EXISTS idx_sales_client_id;
-- SQLite: column left in place
