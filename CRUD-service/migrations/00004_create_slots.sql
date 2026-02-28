-- +goose Up
CREATE TABLE IF NOT EXISTS slots (
    id          SERIAL  PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    appeal_id   INTEGER NOT NULL REFERENCES appeals(id)   ON DELETE CASCADE,
    CONSTRAINT uq_slots_appeal UNIQUE (appeal_id)
);

CREATE INDEX IF NOT EXISTS idx_slots_employee_id ON slots(employee_id);

-- +goose Down
DROP TABLE IF EXISTS slots;
