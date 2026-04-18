package repository

import (
	"fmt"
	"slices"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"crud-service/internal/crud/model"
)

// employeeDB is the database-level representation of Employee.
type employeeDB struct {
	ID           int        `db:"id"`
	Name         string     `db:"name"`
	Surname      string     `db:"surname"`
	Limit        int        `db:"limit"`
	Email        string     `db:"email"`
	Status       string     `db:"status"`
	LastAssignAt *time.Time `db:"last_assign_at"`
	TeamIDs      []int
}

func toEmployeeDB(e model.Employee) employeeDB {
	return employeeDB{
		ID:           e.ID,
		Name:         e.Name,
		Surname:      e.Surname,
		Limit:        e.Limit,
		Email:        e.Email,
		TeamIDs:      e.TeamIDs,
		Status:       e.Status,
		LastAssignAt: e.LastAssignAt,
	}
}

func (e employeeDB) toDomain() model.Employee {
	return model.Employee{
		ID:           e.ID,
		Name:         e.Name,
		Surname:      e.Surname,
		Limit:        e.Limit,
		Email:        e.Email,
		TeamIDs:      e.TeamIDs,
		Status:       e.Status,
		LastAssignAt: e.LastAssignAt,
	}
}

// EmployeeRepository defines CRUD operations for Employee.
type EmployeeRepository interface {
	GetAll() ([]model.Employee, error)
	GetByID(id int) (model.Employee, error)
	Create(tx *sqlx.Tx, e model.Employee) (model.Employee, error)
	Update(tx *sqlx.Tx, id int, e model.Employee) (model.Employee, error)
	Delete(tx *sqlx.Tx, id int) error
	FetchAvailableEmployees(limit int) ([]EmployeeWithAppealsCount, error)
	GetEmployeeActiveAppeals(employeeID int) (int, error)
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
	if err := r.db.Select(&rows, `SELECT id, name, surname, "limit", email, status, last_assign_at FROM employees`); err != nil {
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
	err := r.db.Get(&row, `SELECT id, name, surname, "limit", email, status, last_assign_at FROM employees WHERE id = $1`, id)
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
		`INSERT INTO employees (name, surname, "limit", email, status, last_assign_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		row.Name, row.Surname, row.Limit, row.Email, row.Status, row.LastAssignAt,
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
		`UPDATE employees SET name=$1, surname=$2, "limit"=$3, email=$4, status=$5, last_assign_at=$6 WHERE id=$7`,
		row.Name, row.Surname, row.Limit, row.Email, row.Status, row.LastAssignAt, row.ID,
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

type EmployeeWithAppealsCount struct {
	model.Employee
	ActiveAppealsCount int `db:"active_appeals_count"`
}

func (r *employeeRepo) FetchAvailableEmployees(limit int) ([]EmployeeWithAppealsCount, error) {
	// Упорядоченных список менеджеров для назначения обращений.
	query := `WITH
    active_appeals_count AS (
        SELECT employee_id, COUNT(*) AS active_count
        FROM appeals
        WHERE status='active' AND employee_id IS NOT NULL
        GROUP BY employee_id
    )

SELECT e.id AS id,
       COALESCE(aac.active_count, 0) AS active_appeals_count,
       e.last_assign_at AS last_assign_at,
       COALESCE(array_agg(te.team_id), '{}') AS team_ids
FROM employees e
LEFT JOIN active_appeals_count aac ON aac.employee_id = e.id
LEFT JOIN teams_employees te ON te.employee_id = e.id
WHERE e.status = 'working'
GROUP BY e.id, e.id, aac.active_count, e.last_assign_at
ORDER BY COALESCE(aac.active_count, 0) ASC, COALESCE(e.last_assign_at, '1970-01-01'::timestamp) ASC NULLS FIRST
LIMIT $1
`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("employeeRepo.FetchAvailableManagers: %w", err)
	}
	defer rows.Close()

	var out []EmployeeWithAppealsCount
	for rows.Next() {
		var e EmployeeWithAppealsCount

		teamIDs := pq.Int64Array{}

		if err := rows.Scan(&e.Employee.ID, &e.ActiveAppealsCount, &e.Employee.LastAssignAt, &teamIDs); err != nil {
			return nil, fmt.Errorf("employeeRepo.FetchAvailableManagers: %w", err)
		}

		e.Employee.TeamIDs = make([]int, len(teamIDs))
		for i, id := range teamIDs {
			e.Employee.TeamIDs[i] = int(id)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("employeeRepo.FetchAvailableManagers: %w", err)
	}
	return out, nil
}

func (r *employeeRepo) GetEmployeeActiveAppeals(employeeID int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM appeals WHERE employee_id = $1 AND status = 'active'`, employeeID)
	if err != nil {
		return 0, fmt.Errorf("employeeRepo.GetEmployeeActiveAppeals: %w", err)
	}
	return count, nil
}
