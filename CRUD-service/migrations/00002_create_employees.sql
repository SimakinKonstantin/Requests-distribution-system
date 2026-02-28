-- +goose Up
CREATE TABLE IF NOT EXISTS employees (
    id      SERIAL       PRIMARY KEY,
    name    VARCHAR(255) NOT NULL,
    surname VARCHAR(255) NOT NULL,
    "limit" INTEGER      NOT NULL DEFAULT 0,
    team_id INTEGER      NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS employees;
