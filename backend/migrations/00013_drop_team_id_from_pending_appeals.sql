-- +goose Up
-- +goose StatementBegin
ALTER TABLE pending_appeals
    DROP COLUMN IF EXISTS team_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pending_appeals
    ADD COLUMN IF NOT EXISTS team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE;
-- +goose StatementEnd
