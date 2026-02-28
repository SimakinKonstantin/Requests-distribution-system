-- +goose Up
CREATE TABLE IF NOT EXISTS appeals (
    id          SERIAL  PRIMARY KEY,
    client_id   INTEGER NOT NULL,
    employee_id INTEGER      REFERENCES employees(id) ON DELETE SET NULL,
    theme_id    INTEGER NOT NULL,
    subtheme_id INTEGER NOT NULL REFERENCES subthemes(id) ON DELETE RESTRICT,
    text        TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_appeals_employee_id  ON appeals(employee_id);
CREATE INDEX IF NOT EXISTS idx_appeals_subtheme_id  ON appeals(subtheme_id);
CREATE INDEX IF NOT EXISTS idx_appeals_client_id    ON appeals(client_id);

-- +goose Down
DROP TABLE IF EXISTS appeals;
