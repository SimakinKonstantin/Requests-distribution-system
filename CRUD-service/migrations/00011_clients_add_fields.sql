-- +goose Up
-- +goose StatementBegin
ALTER TABLE clients ADD COLUMN IF NOT EXISTS name TEXT NOT NULL;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS surname TEXT NOT NULL;
ALTER TABLE clients ADD COLUMN IF NOT EXISTS is_vip BOOLEAN NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clients DROP COLUMN IF EXISTS name;
ALTER TABLE clients DROP COLUMN IF EXISTS surname;
ALTER TABLE clients DROP COLUMN IF EXISTS is_vip;
-- +goose StatementEnd
