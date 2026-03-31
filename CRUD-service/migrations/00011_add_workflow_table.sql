-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workflows (
    id      SERIAL       PRIMARY KEY,
    name    TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused'))
    data JSONB NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workflows;
-- +goose StatementEnd
