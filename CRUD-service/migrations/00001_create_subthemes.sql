-- +goose Up
CREATE TABLE IF NOT EXISTS subthemes (
    id   SERIAL      PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS subthemes;
