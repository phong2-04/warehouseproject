ALTER TABLE users
ADD CONSTRAINT chk_users_role
CHECK (role IN ('admin', 'staff'));

CREATE INDEX idx_products_category_id
ON products(category_id);

CREATE INDEX idx_stock_transactions_product_id
ON stock_transactions(product_id);

CREATE INDEX idx_stock_transactions_created_at
ON stock_transactions(created_at);