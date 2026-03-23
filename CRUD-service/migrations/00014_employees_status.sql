-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN IF NOT EXISTS status TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
