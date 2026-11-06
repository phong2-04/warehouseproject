ALTER TABLE stock_transactions
DROP CONSTRAINT IF EXISTS chk_stock_transactions_type;

ALTER TABLE stock_transactions
DROP CONSTRAINT IF EXISTS chk_stock_transactions_quantity_positive;

ALTER TABLE products
DROP CONSTRAINT IF EXISTS chk_products_min_stock_non_negative;

ALTER TABLE products
DROP CONSTRAINT IF EXISTS chk_products_quantity_non_negative;

ALTER TABLE products
DROP CONSTRAINT IF EXISTS chk_products_price_non_negative;
