-- +goose Up
-- +goose StatementBegin
ALTER TABLE slots ADD COLUMN IF NOT EXISTS need_to_remove BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE slots DROP COLUMN IF EXISTS need_to_remove;
-- +goose StatementEnd
