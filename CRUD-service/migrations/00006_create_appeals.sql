-- +goose Up
CREATE TABLE IF NOT EXISTS appeals (
    id          SERIAL  PRIMARY KEY,
    client_id   INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    employee_id INTEGER      REFERENCES employees(id) ON DELETE CASCADE,
    theme_id    INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    subtheme_id INTEGER NOT NULL REFERENCES subthemes(id) ON DELETE CASCADE,
    text        TEXT    NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_appeals_employee_id  ON appeals(employee_id);
CREATE INDEX IF NOT EXISTS idx_appeals_subtheme_id  ON appeals(subtheme_id);
CREATE INDEX IF NOT EXISTS idx_appeals_client_id    ON appeals(client_id);

-- +goose Down
DROP TABLE IF EXISTS appeals;
