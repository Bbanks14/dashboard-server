-- +goose Up
-- +goose StatementBegin
CREATE TABLE regions (
  region_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  manager_id UUID,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE regions
ADD CONSTRAINT fk_manager FOREIGN KEY (manager_id) REFERENCES users(user_id) ON DELETE SET NULL;
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
DROP TRIGGER IF EXISTS trigger_regions_updated_at ON regions;
CREATE TRIGGER trigger_regions_updated_at
BEFORE UPDATE ON regions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE regions DROP CONSTRAINT IF EXISTS fk_manager;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS regions;
-- +goose StatementEnd
