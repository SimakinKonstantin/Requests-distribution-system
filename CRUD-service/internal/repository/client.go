package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// clientDB is the database-level representation of Client.
type clientDB struct {
	ID    int    `db:"id"`
	Email string `db:"email"`
}

func toClientDB(c model.Client) clientDB {
	return clientDB{ID: c.ID, Email: c.Email}
}

func (c clientDB) toDomain() model.Client {
	return model.Client{ID: c.ID, Email: c.Email}
}

// ClientRepository defines CRUD operations for Client.
type ClientRepository interface {
	GetAll() ([]model.Client, error)
	GetByID(id int) (model.Client, error)
	Create(c model.Client) (model.Client, error)
	Update(id int, c model.Client) (model.Client, error)
	Delete(id int) error
}

type clientRepo struct {
	db *sqlx.DB
}

// NewClientRepository returns a PostgreSQL-backed ClientRepository.
func NewClientRepository(db *sqlx.DB) ClientRepository {
	return &clientRepo{db: db}
}

func (r *clientRepo) GetAll() ([]model.Client, error) {
	var rows []clientDB
	if err := r.db.Select(&rows, `SELECT id, email FROM clients ORDER BY id`); err != nil {
		return nil, fmt.Errorf("clientRepo.GetAll: %w", err)
	}
	result := make([]model.Client, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *clientRepo) GetByID(id int) (model.Client, error) {
	var row clientDB
	if err := r.db.Get(&row, `SELECT id, email FROM clients WHERE id = $1`, id); err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Create(c model.Client) (model.Client, error) {
	row := toClientDB(c)
	err := r.db.QueryRowx(
		`INSERT INTO clients (email) VALUES ($1) RETURNING id`,
		row.Email,
	).Scan(&row.ID)
	if err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Update(id int, c model.Client) (model.Client, error) {
	row := toClientDB(c)
	row.ID = id
	res, err := r.db.Exec(
		`UPDATE clients SET email=$1 WHERE id=$2`,
		row.Email, row.ID,
	)
	if err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Client{}, err
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM clients WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("clientRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}
