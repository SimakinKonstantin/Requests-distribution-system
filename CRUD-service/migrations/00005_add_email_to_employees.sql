-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees
    ADD COLUMN email VARCHAR(255) NOT NULL DEFAULT '' UNIQUE;
ALTER TABLE employees
    ALTER COLUMN email DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN email;
-- +goose StatementEnd
