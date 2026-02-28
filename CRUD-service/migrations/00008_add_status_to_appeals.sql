-- +goose Up
-- +goose StatementBegin
ALTER TABLE appeals
    ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'closed'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE appeals DROP COLUMN status;
-- +goose StatementEnd
