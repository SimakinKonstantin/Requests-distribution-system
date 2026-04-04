package repository

import (
	"crud-service/internal/crud/model"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// appealDB is the database-level representation of Appeal.
// EmployeeID, SubthemeID and TeamID are nullable.
type appealDB struct {
	ID         int           `db:"id"`
	ClientID   int           `db:"client_id"`
	EmployeeID sql.NullInt64 `db:"employee_id"`
	ThemeID    int           `db:"theme_id"`
	SubthemeID sql.NullInt64 `db:"subtheme_id"`
	Text       string        `db:"text"`
	Status     string        `db:"status"`
	TeamID     sql.NullInt64 `db:"team_id"`
}

// toAppealDB converts a domain Appeal to the DB struct.
// EmployeeID, SubthemeID and TeamID are pointers — nil is stored as NULL.
func toAppealDB(a model.Appeal) appealDB {
	empID := sql.NullInt64{}
	if a.EmployeeID != nil {
		empID = sql.NullInt64{Int64: int64(*a.EmployeeID), Valid: true}
	}
	subthemeID := sql.NullInt64{}
	if a.SubthemeID != nil {
		subthemeID = sql.NullInt64{Int64: int64(*a.SubthemeID), Valid: true}
	}
	teamID := sql.NullInt64{}
	if a.TeamID != nil {
		teamID = sql.NullInt64{Int64: int64(*a.TeamID), Valid: true}
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
		SubthemeID: subthemeID,
		Text:       a.Text,
		Status:     status,
		TeamID:     teamID,
	}
}

func (a appealDB) toDomain() model.Appeal {
	var empID *int
	if a.EmployeeID.Valid {
		v := int(a.EmployeeID.Int64)
		empID = &v
	}
	var subthemeID *int
	if a.SubthemeID.Valid {
		v := int(a.SubthemeID.Int64)
		subthemeID = &v
	}
	var teamID *int
	if a.TeamID.Valid {
		v := int(a.TeamID.Int64)
		teamID = &v
	}
	return model.Appeal{
		ID:         a.ID,
		ClientID:   a.ClientID,
		EmployeeID: empID,
		ThemeID:    a.ThemeID,
		SubthemeID: subthemeID,
		Text:       a.Text,
		Status:     a.Status,
		TeamID:     teamID,
	}
}

// AppealRepository defines CRUD operations for Appeal.
type AppealRepository interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(tx *sqlx.Tx, a model.Appeal) (model.Appeal, error)
	Update(tx *sqlx.Tx, id int, a model.Appeal) (model.Appeal, error)
	Delete(tx *sqlx.Tx, id int) error
	Close(tx *sqlx.Tx, id int) (model.Appeal, error)
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
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text, status, team_id
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
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text, status, team_id
		 FROM appeals WHERE id = $1`, id,
	)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.GetByID: %w", err)
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Create(tx *sqlx.Tx, a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)

	err := tx.QueryRowx(
		`INSERT INTO appeals (client_id, employee_id, theme_id, subtheme_id, text, status, team_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		row.ClientID, row.EmployeeID, row.ThemeID, row.SubthemeID, row.Text, row.Status, row.TeamID,
	).Scan(&row.ID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Update(tx *sqlx.Tx, id int, a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)
	row.ID = id
	res, err := tx.Exec(
		`UPDATE appeals
		 SET client_id=$1, employee_id=$2, theme_id=$3, subtheme_id=$4, text=$5, status=$6, team_id=$7
		 WHERE id=$8`,
		row.ClientID, row.EmployeeID, row.ThemeID, row.SubthemeID, row.Text, row.Status, row.TeamID, row.ID,
	)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Appeal{}, err
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM appeals WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("appealRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

func (r *appealRepo) Close(tx *sqlx.Tx, id int) (model.Appeal, error) {
	var row appealDB
	err := tx.QueryRowx(
		`UPDATE appeals SET status='closed' WHERE id=$1
		 RETURNING id, client_id, employee_id, theme_id, subtheme_id, text, status, team_id`,
		id,
	).StructScan(&row)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Close: %w", err)
	}
	return row.toDomain(), nil
}
