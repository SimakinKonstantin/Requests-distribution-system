package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// themeDB is the database-level representation of Theme.
type themeDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func toThemeDB(t model.Theme) themeDB {
	return themeDB{ID: t.ID, Name: t.Name}
}

func (t themeDB) toDomain() model.Theme {
	return model.Theme{ID: t.ID, Name: t.Name}
}

// ThemeRepository defines CRUD operations for Theme.
type ThemeRepository interface {
	GetAll() ([]model.Theme, error)
	GetByID(id int) (model.Theme, error)
	Create(tx *sqlx.Tx, t model.Theme) (model.Theme, error)
	Update(tx *sqlx.Tx, id int, t model.Theme) (model.Theme, error)
	Delete(tx *sqlx.Tx, id int) error
}

type themeRepo struct {
	db *sqlx.DB
}

// NewThemeRepository returns a PostgreSQL-backed ThemeRepository.
func NewThemeRepository(db *sqlx.DB) ThemeRepository {
	return &themeRepo{db: db}
}

func (r *themeRepo) GetAll() ([]model.Theme, error) {
	var rows []themeDB
	if err := r.db.Select(&rows, `SELECT id, name FROM themes ORDER BY id`); err != nil {
		return nil, fmt.Errorf("themeRepo.GetAll: %w", err)
	}
	result := make([]model.Theme, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *themeRepo) GetByID(id int) (model.Theme, error) {
	var row themeDB
	if err := r.db.Get(&row, `SELECT id, name FROM themes WHERE id = $1`, id); err != nil {
		return model.Theme{}, fmt.Errorf("themeRepo.GetByID: %w", err)
	}
	return row.toDomain(), nil
}

func (r *themeRepo) Create(tx *sqlx.Tx, t model.Theme) (model.Theme, error) {
	row := toThemeDB(t)
	err := tx.QueryRowx(
		`INSERT INTO themes (name) VALUES ($1) RETURNING id`,
		row.Name,
	).Scan(&row.ID)
	if err != nil {
		return model.Theme{}, fmt.Errorf("themeRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *themeRepo) Update(tx *sqlx.Tx, id int, t model.Theme) (model.Theme, error) {
	row := toThemeDB(t)
	row.ID = id
	res, err := tx.Exec(
		`UPDATE themes SET name=$1 WHERE id=$2`,
		row.Name, row.ID,
	)
	if err != nil {
		return model.Theme{}, fmt.Errorf("themeRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Theme{}, err
	}
	return row.toDomain(), nil
}

func (r *themeRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM themes WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("themeRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}
