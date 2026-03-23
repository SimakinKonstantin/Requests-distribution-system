package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// subthemeDB is the database-level representation of Subtheme.
type subthemeDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func toSubthemeDB(s model.Subtheme) subthemeDB {
	return subthemeDB{
		ID:   s.ID,
		Name: s.Name,
	}
}

func (s subthemeDB) toDomain() model.Subtheme {
	return model.Subtheme{
		ID:   s.ID,
		Name: s.Name,
	}
}

// SubthemeRepository defines CRUD operations for Subtheme.
type SubthemeRepository interface {
	GetAll() ([]model.Subtheme, error)
	GetByID(id int) (model.Subtheme, error)
	Create(s model.Subtheme) (model.Subtheme, error)
	Update(id int, s model.Subtheme) (model.Subtheme, error)
	Delete(id int) error
}

type subthemeRepo struct {
	db *sqlx.DB
}

// NewSubthemeRepository returns a PostgreSQL-backed SubthemeRepository.
func NewSubthemeRepository(db *sqlx.DB) SubthemeRepository {
	return &subthemeRepo{db: db}
}

func (r *subthemeRepo) GetAll() ([]model.Subtheme, error) {
	var rows []subthemeDB
	if err := r.db.Select(&rows, `SELECT id, name, theme_id FROM subthemes`); err != nil {
		return nil, fmt.Errorf("subthemeRepo.GetAll: %w", err)
	}

	result := make([]model.Subtheme, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *subthemeRepo) GetByID(id int) (model.Subtheme, error) {
	var row subthemeDB
	err := r.db.Get(&row, `SELECT id, name, theme_id FROM subthemes WHERE id = $1`, id)
	if err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *subthemeRepo) Create(s model.Subtheme) (model.Subtheme, error) {
	row := toSubthemeDB(s)
	err := r.db.QueryRowx(
		`INSERT INTO subthemes (name) VALUES ($1) RETURNING id`,
		row.Name,
	).Scan(&row.ID)
	if err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *subthemeRepo) Update(id int, s model.Subtheme) (model.Subtheme, error) {
	row := toSubthemeDB(s)
	row.ID = id
	res, err := r.db.Exec(`UPDATE subthemes SET name=$1, WHERE id=$2`, row.Name, row.ID)
	if err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Subtheme{}, err
	}
	return row.toDomain(), nil
}

func (r *subthemeRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM subthemes WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("subthemeRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}
