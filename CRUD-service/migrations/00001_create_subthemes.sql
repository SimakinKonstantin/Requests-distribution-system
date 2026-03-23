-- +goose Up
CREATE TABLE IF NOT EXISTS subthemes (
    id   SERIAL      PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS subthemes;
