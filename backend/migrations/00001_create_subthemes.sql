-- +goose Up
CREATE TABLE IF NOT EXISTS subthemes (
    id   SERIAL      PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS subthemes;
