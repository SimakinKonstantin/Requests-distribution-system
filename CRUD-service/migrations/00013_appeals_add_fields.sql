-- +goose Up
-- +goose StatementBegin
ALTER TABLE appeals ADD COLUMN IF NOT EXISTS status TEXT NOT NULL;
ALTER TABLE appeals ADD COLUMN IF NOT EXISTS team_id INTEGER NOT NULL;

ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_team;
ALTER TABLE appeals ADD CONSTRAINT fk_appeals_team FOREIGN KEY (team_id) REFERENCES teams(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE appeals DROP COLUMN IF EXISTS status;
ALTER TABLE appeals DROP COLUMN IF EXISTS team_id;
ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_team;
-- +goose StatementEnd
