-- +goose Up
-- +goose StatementBegin
CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    sale_id UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    order_id VARCHAR(255) NOT NULL UNIQUE,
    product_id UUID,
    customer_id UUID,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price >= 0),
    total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount >= 0),
    status VARCHAR(20) NOT NULL CHECK (status IN ('Success', 'Pending')),
    category VARCHAR(100),
    sale_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Create indexes for performances
CREATE INDEX idx_sales_order_id ON sales (order_id);
CREATE INDEX idx_sales_sale_date ON sales (sale_date);
CREATE INDEX idx_sales_category ON sales (category);
CREATE INDEX idx_sales_product_id ON sales (product_id);
CREATE INDEX idx_sales_customer_id ON sales (customer_id);

-- +goose StatementBegin
ALTER TABLE sales
ADD CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales
ADD CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trigger_sales_updated_at
  BEFORE UPDATE ON sales
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_sales_updated_at ON sales;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_updated_at_column;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_product;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS sales;
-- +goose StatementEnd
