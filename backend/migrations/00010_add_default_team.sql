-- +goose Up
-- +goose StatementBegin
INSERT INTO teams (name) VALUES ('Не распределенные');
INSERT INTO teams (name) VALUES ('Не распределенные VIP');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM teams WHERE name = 'Не распределенные';
DELETE FROM teams WHERE name = 'Не распределенные VIP';
-- +goose StatementEnd
