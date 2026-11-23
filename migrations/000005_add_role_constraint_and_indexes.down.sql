DROP INDEX IF EXISTS idx_stock_transactions_created_at;
DROP INDEX IF EXISTS idx_stock_transactions_product_id;
DROP INDEX IF EXISTS idx_products_category_id;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS chk_users_role;