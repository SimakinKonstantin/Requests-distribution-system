-- +goose Up
CREATE TABLE IF NOT EXISTS employees (
    id      SERIAL       PRIMARY KEY,
    name    TEXT NOT NULL,
    surname TEXT NOT NULL,
    "limit" INTEGER      NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS employees;
