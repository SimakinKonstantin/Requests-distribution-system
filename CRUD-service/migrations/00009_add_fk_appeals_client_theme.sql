-- +goose Up
-- +goose StatementBegin
ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_client;
ALTER TABLE appeals ADD CONSTRAINT fk_appeals_client FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE;

ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_theme;
ALTER TABLE appeals ADD CONSTRAINT fk_appeals_theme FOREIGN KEY (theme_id) REFERENCES themes(id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_client;
ALTER TABLE appeals DROP CONSTRAINT IF EXISTS fk_appeals_theme;
-- +goose StatementEnd
