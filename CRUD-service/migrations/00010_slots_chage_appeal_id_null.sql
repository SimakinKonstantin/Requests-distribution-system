-- +goose Up
-- +goose StatementBegin
ALTER TABLE slots ALTER COLUMN appeal_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE slots ALTER COLUMN appeal_id SET NOT NULL;
-- +goose StatementEnd
