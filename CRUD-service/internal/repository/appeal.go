package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// appealDB is the database-level representation of Appeal.
// employee_id is nullable — an appeal may not yet have an assigned employee.
type appealDB struct {
	ID         int           `db:"id"`
	ClientID   int           `db:"client_id"`
	EmployeeID sql.NullInt64 `db:"employee_id"`
	ThemeID    int           `db:"theme_id"`
	SubthemeID int           `db:"subtheme_id"`
	Text       string        `db:"text"`
	Status     string        `db:"status"`
}

// toAppealDB converts a domain Appeal to the DB struct.
// EmployeeID == 0 is treated as "not assigned" and stored as NULL.
func toAppealDB(a model.Appeal) appealDB {
	empID := sql.NullInt64{}
	if a.EmployeeID != 0 {
		empID = sql.NullInt64{Int64: int64(a.EmployeeID), Valid: true}
	}
	status := a.Status
	if status == "" {
		status = "active"
	}
	return appealDB{
		ID:         a.ID,
		ClientID:   a.ClientID,
		EmployeeID: empID,
		ThemeID:    a.ThemeID,
		SubthemeID: a.SubthemeID,
		Text:       a.Text,
		Status:     status,
	}
}

func (a appealDB) toDomain() model.Appeal {
	empID := 0
	if a.EmployeeID.Valid {
		empID = int(a.EmployeeID.Int64)
	}
	return model.Appeal{
		ID:         a.ID,
		ClientID:   a.ClientID,
		EmployeeID: empID,
		ThemeID:    a.ThemeID,
		SubthemeID: a.SubthemeID,
		Text:       a.Text,
		Status:     a.Status,
	}
}

// AppealRepository defines CRUD operations for Appeal.
type AppealRepository interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(a model.Appeal) (model.Appeal, error)
	Update(id int, a model.Appeal) (model.Appeal, error)
	Delete(id int) error
	Close(id int) (model.Appeal, error)
}

type appealRepo struct {
	db *sqlx.DB
}

// NewAppealRepository returns a PostgreSQL-backed AppealRepository.
func NewAppealRepository(db *sqlx.DB) AppealRepository {
	return &appealRepo{db: db}
}

func (r *appealRepo) GetAll() ([]model.Appeal, error) {
	var rows []appealDB
	if err := r.db.Select(&rows,
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text, status
		 FROM appeals ORDER BY id`,
	); err != nil {
		return nil, fmt.Errorf("appealRepo.GetAll: %w", err)
	}

	result := make([]model.Appeal, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *appealRepo) GetByID(id int) (model.Appeal, error) {
	var row appealDB
	err := r.db.Get(&row,
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text, status
		 FROM appeals WHERE id = $1`, id,
	)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Create(a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)
	var empArg interface{}
	if row.EmployeeID.Valid {
		empArg = row.EmployeeID.Int64
	}
	err := r.db.QueryRowx(
		`INSERT INTO appeals (client_id, employee_id, theme_id, subtheme_id, text, status)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		row.ClientID, empArg, row.ThemeID, row.SubthemeID, row.Text, row.Status,
	).Scan(&row.ID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Update(id int, a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)
	row.ID = id
	var empArg interface{}
	if row.EmployeeID.Valid {
		empArg = row.EmployeeID.Int64
	}
	res, err := r.db.Exec(
		`UPDATE appeals
		 SET client_id=$1, employee_id=$2, theme_id=$3, subtheme_id=$4, text=$5, status=$6
		 WHERE id=$7`,
		row.ClientID, empArg, row.ThemeID, row.SubthemeID, row.Text, row.Status, row.ID,
	)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Appeal{}, err
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM appeals WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("appealRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

// Close sets the appeal status to "closed" and returns the updated appeal.
func (r *appealRepo) Close(id int) (model.Appeal, error) {
	var row appealDB
	err := r.db.QueryRowx(
		`UPDATE appeals SET status='closed' WHERE id=$1
		 RETURNING id, client_id, employee_id, theme_id, subtheme_id, text, status`,
		id,
	).StructScan(&row)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Close: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}
