package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// employeeDB is the database-level representation of Employee.
type employeeDB struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	Surname string `db:"surname"`
	Limit   int    `db:"limit"`
	TeamID  int    `db:"team_id"`
}

func toEmployeeDB(e model.Employee) employeeDB {
	return employeeDB{
		ID:      e.ID,
		Name:    e.Name,
		Surname: e.Surname,
		Limit:   e.Limit,
		TeamID:  e.TeamID,
	}
}

func (e employeeDB) toDomain() model.Employee {
	return model.Employee{
		ID:      e.ID,
		Name:    e.Name,
		Surname: e.Surname,
		Limit:   e.Limit,
		TeamID:  e.TeamID,
	}
}

// EmployeeRepository defines CRUD operations for Employee.
type EmployeeRepository interface {
	GetAll() ([]model.Employee, error)
	GetByID(id int) (model.Employee, error)
	Create(e model.Employee) (model.Employee, error)
	Update(id int, e model.Employee) (model.Employee, error)
	Delete(id int) error
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
	if err := r.db.Select(&rows, `SELECT id, name, surname, "limit", team_id FROM employees`); err != nil {
		return nil, fmt.Errorf("employeeRepo.GetAll: %w", err)
	}

	result := make([]model.Employee, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *employeeRepo) GetByID(id int) (model.Employee, error) {
	var row employeeDB
	err := r.db.Get(&row, `SELECT id, name, surname, "limit", team_id FROM employees WHERE id = $1`, id)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *employeeRepo) Create(e model.Employee) (model.Employee, error) {
	row := toEmployeeDB(e)
	err := r.db.QueryRowx(
		`INSERT INTO employees (name, surname, "limit", team_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		row.Name, row.Surname, row.Limit, row.TeamID,
	).Scan(&row.ID)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *employeeRepo) Update(id int, e model.Employee) (model.Employee, error) {
	row := toEmployeeDB(e)
	row.ID = id
	res, err := r.db.Exec(
		`UPDATE employees SET name=$1, surname=$2, "limit"=$3, team_id=$4 WHERE id=$5`,
		row.Name, row.Surname, row.Limit, row.TeamID, row.ID,
	)
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Employee{}, err
	}
	return row.toDomain(), nil
}

func (r *employeeRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM employees WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("employeeRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}
