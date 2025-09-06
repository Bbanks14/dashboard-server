-- +goose Up
-- +goose StatementBegin
CREATE TABLE forecasts_table (
    forecast_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id UUID,
    product_id UUID,
    region_id UUID,
    forecasted_amount DECIMAL(10, 2) NOT NULL CHECK (forecasted_amount >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table
ADD CONSTRAINT fk_period_id FOREIGN KEY (period_id) REFERENCES periods(period_id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table
ADD CONSTRAINT fk_product_id FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table
ADD CONSTRAINT fk_region_id FOREIGN KEY (region_id) REFERENCES regions(region_id) ON DELETE SET NULL;
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

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_forecasts_updated_at ON forecasts;
CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON forecasts
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table DROP CONSTRAINT IF EXISTS fk_period_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table DROP CONSTRAINT IF EXISTS fk_product_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE forecasts_table DROP CONSTRAINT IF EXISTS fk_region_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS forecasts_table;
-- +goose StatementEnd
