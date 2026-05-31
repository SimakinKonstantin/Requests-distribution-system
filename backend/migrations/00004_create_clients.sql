-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clients (
    id    SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    is_vip BOOLEAN NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clients;
-- +goose StatementEnd
