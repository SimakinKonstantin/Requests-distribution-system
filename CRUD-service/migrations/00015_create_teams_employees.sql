-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams_employees (
    team_id INTEGER NOT NULL,
    employee_id INTEGER NOT NULL,
    PRIMARY KEY (team_id, employee_id)
);

ALTER TABLE teams_employees DROP CONSTRAINT IF EXISTS fk_teams_employees_team;
ALTER TABLE teams_employees ADD CONSTRAINT fk_teams_employees_team FOREIGN KEY (team_id) REFERENCES teams(id);

ALTER TABLE teams_employees DROP CONSTRAINT IF EXISTS fk_teams_employees_employee;
ALTER TABLE teams_employees ADD CONSTRAINT fk_teams_employees_employee FOREIGN KEY (employee_id) REFERENCES employees(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP CONSTRAINT IF EXISTS fk_teams_employees_employee;
DROP CONSTRAINT IF EXISTS fk_teams_employees_team;
DROP TABLE IF EXISTS teams_employees;
-- +goose StatementEnd
