-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams_employees (
    team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, employee_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teams_employees;
-- +goose StatementEnd
