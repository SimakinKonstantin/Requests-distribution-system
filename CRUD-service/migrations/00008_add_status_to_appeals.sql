-- +goose Up
-- +goose StatementBegin
ALTER TABLE appeals
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'closed'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE appeals DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
