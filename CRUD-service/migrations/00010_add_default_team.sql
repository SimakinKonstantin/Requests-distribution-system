-- +goose Up
-- +goose StatementBegin
INSERT INTO teams (name) VALUES ('Не распределенные');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE slots DROP COLUMN IF EXISTS need_to_remove;
-- +goose StatementEnd
