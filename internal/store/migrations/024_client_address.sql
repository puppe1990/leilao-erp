-- up
-- Client address + legal/person type, filled by CNPJ lookup (BrasilAPI).
-- SQLite: ALTER ADD COLUMN is fine; migrations run once.
ALTER TABLE clients ADD COLUMN client_type TEXT NOT NULL DEFAULT 'person';
ALTER TABLE clients ADD COLUMN cep TEXT;
ALTER TABLE clients ADD COLUMN street TEXT;
ALTER TABLE clients ADD COLUMN number TEXT;
ALTER TABLE clients ADD COLUMN complement TEXT;
ALTER TABLE clients ADD COLUMN neighborhood TEXT;
ALTER TABLE clients ADD COLUMN city TEXT;
ALTER TABLE clients ADD COLUMN state TEXT;

-- down
-- SQLite cannot DROP COLUMN reliably; leave columns if rolled back manually.
