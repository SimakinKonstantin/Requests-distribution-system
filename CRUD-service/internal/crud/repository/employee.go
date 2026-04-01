package repository

import (
	"fmt"
	"slices"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/crud/model"
)

// employeeDB is the database-level representation of Employee.
type employeeDB struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	Surname string `db:"surname"`
	Limit   int    `db:"limit"`
	Email   string `db:"email"`
	Status  string `db:"status"`
	TeamIDs []int
}

func toEmployeeDB(e model.Employee) employeeDB {
	return employeeDB{
		ID:      e.ID,
		Name:    e.Name,
		Surname: e.Surname,
		Limit:   e.Limit,
		Email:   e.Email,
		TeamIDs: e.TeamIDs,
		Status:  e.Status,
	}
}

func (e employeeDB) toDomain() model.Employee {
	return model.Employee{
		ID:      e.ID,
		Name:    e.Name,
		Surname: e.Surname,
		Limit:   e.Limit,
		Email:   e.Email,
		TeamIDs: e.TeamIDs,
		Status:  e.Status,
	}
}

// EmployeeRepository defines CRUD operations for Employee.
type EmployeeRepository interface {
	GetAll() ([]model.Employee, error)
	GetByID(id int) (model.Employee, error)
	Create(tx *sqlx.Tx, e model.Employee) (model.Employee, error)
	Update(tx *sqlx.Tx, id int, e model.Employee) (model.Employee, error)
	Delete(tx *sqlx.Tx, id int) error
}

type employeeRepo struct {
	db *sqlx.DB
}

// NewEmployeeRepository returns a PostgreSQL-backed EmployeeRepository.
func NewEmployeeRepository(db *sqlx.DB) EmployeeRepository {
	return &employeeRepo{db: db}
}

func (r *employeeRepo) GetAll() ([]model.Employee, error) {
	var rows []employeeDB
	if err := r.db.Select(&rows, `SELECT id, name, surname, "limit", email, status FROM employees`); err != nil {
		return nil, fmt.Errorf("employeeRepo.GetAll: %w", err)
	}

	result := make([]model.Employee, len(rows))
	for i, row := range rows {
		filledRow, err := r.fillTeams(row)
		if err != nil {
			return nil, fmt.Errorf("employeeRepo.GetAll fillTeams: %w", err)
		}
		result[i] = filledRow.toDomain()
	}
	return result, nil
}

func (r *employeeRepo) GetByID(id int) (model.Employee, error) {
	var row employeeDB
	err := r.db.Get(&row, `SELECT id, name, surname, "limit", email, status FROM employees WHERE id = $1`, id)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.GetByID: %w", err)
	}

	filledRow, err := r.fillTeams(row)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.GetByID fillTeams: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *employeeRepo) Create(tx *sqlx.Tx, e model.Employee) (model.Employee, error) {
	row := toEmployeeDB(e)
	err := tx.QueryRowx(
		`INSERT INTO employees (name, surname, "limit", email, status) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		row.Name, row.Surname, row.Limit, row.Email, row.Status,
	).Scan(&row.ID)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.Create: %w", err)
	}

	for _, teamID := range row.TeamIDs {
		_, err = tx.Exec(
			`INSERT INTO teams_employees (employee_id, team_id) VALUES ($1, $2)`,
			row.ID, teamID,
		)
		if err != nil {
			return model.Employee{}, fmt.Errorf("employeeRepo.Create teams_employees: %w", err)
		}
	}

	return row.toDomain(), nil
}

func (r *employeeRepo) Update(tx *sqlx.Tx, id int, e model.Employee) (model.Employee, error) {
	var err error

	row := toEmployeeDB(e)
	row.ID = id
	_, err = tx.Exec(
		`UPDATE employees SET name=$1, surname=$2, "limit"=$3, email=$4, status=$5 WHERE id=$6`,
		row.Name, row.Surname, row.Limit, row.Email, row.Status, row.ID,
	)
	if err != nil {
		rberr := tx.Rollback()
		return model.Employee{}, fmt.Errorf("employeeRepo.Update: %w, rollback error: %w", err, rberr)
	}

	type TeamEmployee struct {
		TeamID int `db:"team_id"`
	}

	var currentTeams []TeamEmployee
	err = tx.Select(&currentTeams, `SELECT team_id FROM teams_employees WHERE employee_id = $1`, row.ID)
	if err != nil {
		rberr := tx.Rollback()
		return model.Employee{}, fmt.Errorf("employeeRepo.Update teams_employees: %w, rollback error: %w", err, rberr)
	}

	for _, currentTeam := range currentTeams {
		if !slices.Contains(row.TeamIDs, currentTeam.TeamID) {
			_, err = tx.Exec(
				`DELETE FROM teams_employees (employee_id, team_id) VALUES ($1, $2)`,
				row.ID, currentTeam.TeamID,
			)
			if err != nil {
				rberr := tx.Rollback()
				return model.Employee{}, fmt.Errorf("employeeRepo.Update teams_employees: %w, rollback error: %w", err, rberr)
			}
		}
	}

	for _, newTeam := range row.TeamIDs {
		if !slices.Contains(currentTeams, TeamEmployee{TeamID: newTeam}) {
			_, err = tx.Exec(
				`INSERT INTO teams_employees (employee_id, team_id) VALUES ($1, $2)`,
				row.ID, newTeam,
			)
			if err != nil {
				rberr := tx.Rollback()
				return model.Employee{}, fmt.Errorf("employeeRepo.Update teams_employees: %w, rollback error: %w", err, rberr)
			}
		}
	}

	filledRow, err := r.fillTeams(row)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.Update fillTeams: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *employeeRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM employees WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("employeeRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

func (r *employeeRepo) fillTeams(e employeeDB) (employeeDB, error) {
	var teams []int
	err := r.db.Select(&teams, `SELECT team_id FROM teams_employees WHERE employee_id = $1`, e.ID)
	if err != nil {
		return employeeDB{}, fmt.Errorf("employeeRepo.fillTeams teams_employees: %w", err)
	}
	e.TeamIDs = teams
	return e, nil
}
