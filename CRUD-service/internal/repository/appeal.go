package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// appealDB is the database-level representation of Appeal.
type appealDB struct {
	ID         int    `db:"id"`
	ClientID   int    `db:"client_id"`
	EmployeeID int    `db:"employee_id"`
	ThemeID    int    `db:"theme_id"`
	SubthemeID int    `db:"subtheme_id"`
	Text       string `db:"text"`
}

func toAppealDB(a model.Appeal) appealDB {
	return appealDB{
		ID:         a.ID,
		ClientID:   a.ClientID,
		EmployeeID: a.EmployeeID,
		ThemeID:    a.ThemeID,
		SubthemeID: a.SubthemeID,
		Text:       a.Text,
	}
}

func (a appealDB) toDomain() model.Appeal {
	return model.Appeal{
		ID:         a.ID,
		ClientID:   a.ClientID,
		EmployeeID: a.EmployeeID,
		ThemeID:    a.ThemeID,
		SubthemeID: a.SubthemeID,
		Text:       a.Text,
	}
}

// AppealRepository defines CRUD operations for Appeal.
type AppealRepository interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(a model.Appeal) (model.Appeal, error)
	Update(id int, a model.Appeal) (model.Appeal, error)
	Delete(id int) error
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
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text FROM appeals`,
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
		`SELECT id, client_id, employee_id, theme_id, subtheme_id, text FROM appeals WHERE id = $1`, id,
	)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Create(a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)
	err := r.db.QueryRowx(
		`INSERT INTO appeals (client_id, employee_id, theme_id, subtheme_id, text)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		row.ClientID, row.EmployeeID, row.ThemeID, row.SubthemeID, row.Text,
	).Scan(&row.ID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *appealRepo) Update(id int, a model.Appeal) (model.Appeal, error) {
	row := toAppealDB(a)
	row.ID = id
	res, err := r.db.Exec(
		`UPDATE appeals SET client_id=$1, employee_id=$2, theme_id=$3, subtheme_id=$4, text=$5 WHERE id=$6`,
		row.ClientID, row.EmployeeID, row.ThemeID, row.SubthemeID, row.Text, row.ID,
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
