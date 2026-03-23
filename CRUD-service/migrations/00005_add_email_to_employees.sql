-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN IF NOT EXISTS email TEXT UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN IF EXISTS email;
-- +goose StatementEnd
