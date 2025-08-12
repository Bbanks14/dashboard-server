-- +goose Up
-- +goose StatementBegin
CREATE TABLE sales_targets (
    target_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    region_id UUID,
    period_id UUID,
    target_amount DECIMAL(10, 2) NOT NULL CHECK (target_amount >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets
ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets
ADD CONSTRAINT fk_region_id FOREIGN KEY (region_id) REFERENCES regions(region_id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets
ADD CONSTRAINT fk_period_id FOREIGN KEY (period_id) REFERENCES periods(period_id) ON DELETE SET NULL;
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
DROP TRIGGER IF EXISTS trigger_targets_updated_at ON targets;
CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON targets
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets DROP CONSTRAINT IF EXISTS fk_user_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets DROP CONSTRAINT IF EXISTS fk_region_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE sales_targets DROP CONSTRAINT IF EXISTS fk_period_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS sales_targets;
-- +goose StatementEnd
