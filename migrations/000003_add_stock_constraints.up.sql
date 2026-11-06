ALTER TABLE products
ADD CONSTRAINT chk_products_price_non_negative
CHECK (price >= 0);

ALTER TABLE products
ADD CONSTRAINT chk_products_quantity_non_negative
CHECK (quantity >= 0);

ALTER TABLE products
ADD CONSTRAINT chk_products_min_stock_non_negative
CHECK (min_stock >= 0);

ALTER TABLE stock_transactions
ADD CONSTRAINT chk_stock_transactions_quantity_positive
CHECK (quantity > 0);

ALTER TABLE stock_transactions
ADD CONSTRAINT chk_stock_transactions_type
CHECK (transaction_type IN ('import', 'export'));
