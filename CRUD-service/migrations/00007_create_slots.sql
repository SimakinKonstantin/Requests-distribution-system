-- +goose Up
CREATE TABLE IF NOT EXISTS slots (
    id          SERIAL  PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    appeal_id   INTEGER REFERENCES appeals(id) ON DELETE CASCADE UNIQUE,
    need_to_remove BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_slots_employee_id ON slots(employee_id);

-- +goose Down
DROP TABLE IF EXISTS slots;
