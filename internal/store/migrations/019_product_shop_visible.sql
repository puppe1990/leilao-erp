-- up
ALTER TABLE products ADD COLUMN shop_visible INTEGER NOT NULL DEFAULT 0;

-- Products already eligible for the public catalog keep appearing.
UPDATE products
SET shop_visible = 1
WHERE EXISTS (
  SELECT 1 FROM product_media m
  WHERE m.product_id = products.id AND m.kind = 'photo'
)
AND EXISTS (
  SELECT 1 FROM items i
  WHERE i.product_id = products.id AND i.status = 'in_stock'
);

-- down
-- SQLite: column left in place
