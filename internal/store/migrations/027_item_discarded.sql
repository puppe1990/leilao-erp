-- up
-- Baixa de unidade: tira o item do estoque sem apagar histórico nem custo.
-- status continua com os 3 valores originais; discarded_at marca a baixa.
ALTER TABLE items ADD COLUMN discarded_at DATETIME;

CREATE INDEX IF NOT EXISTS idx_items_discarded ON items(discarded_at);

-- down
DROP INDEX IF EXISTS idx_items_discarded;
-- SQLite: column left in place
