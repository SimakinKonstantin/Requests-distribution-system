-- +goose Up
-- +goose StatementBegin
ALTER TABLE appeals
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Добавляем поля для балансировщика в таблицу employees
ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS last_assign_at        TIMESTAMP DEFAULT NULL;

-- Очередь ожидающих распределения обращений
CREATE TABLE IF NOT EXISTS pending_appeals (
    appeal_id  INTEGER PRIMARY KEY REFERENCES appeals(id) ON DELETE CASCADE,
    team_id    INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    priority   INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- todo Check indexes
-- CREATE INDEX IF NOT EXISTS pending_appeals_priority_idx
--     ON pending_appeals (priority DESC, updated_at ASC);

-- CREATE INDEX IF NOT EXISTS slots_free_idx
--     ON slots (employee_id, updated_at) WHERE appeal_id IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- DROP INDEX IF EXISTS slots_free_idx;
-- DROP INDEX IF EXISTS pending_appeals_priority_idx;
DROP TABLE IF EXISTS pending_appeals;

ALTER TABLE employees
    DROP COLUMN IF EXISTS last_assign_at,

ALTER TABLE appeals
    DROP COLUMN IF EXISTS created_at,

-- +goose StatementEnd
