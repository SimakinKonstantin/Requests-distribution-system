package repository

import (
	"crud-service/internal/crud/model"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type clientDB struct {
	ID      int    `db:"id"`
	Email   string `db:"email"`
	Name    string `db:"name"`
	Surname string `db:"surname"`
	IsVIP   bool   `db:"is_vip"`
}

func toClientDB(c model.Client) clientDB {
	return clientDB{ID: c.ID, Email: c.Email, Name: c.Name, Surname: c.Surname, IsVIP: c.IsVIP}
}

func (c clientDB) toDomain() model.Client {
	return model.Client{ID: c.ID, Email: c.Email, Name: c.Name, Surname: c.Surname, IsVIP: c.IsVIP}
}

type ClientRepository interface {
	GetAll() ([]model.Client, error)
	GetByID(id int) (model.Client, error)
	Create(tx *sqlx.Tx, c model.Client) (model.Client, error)
	Update(tx *sqlx.Tx, id int, c model.Client) (model.Client, error)
	Delete(tx *sqlx.Tx, id int) error
	GetEmails() ([]string, error)
}

type clientRepo struct {
	db *sqlx.DB
}

func NewClientRepository(db *sqlx.DB) ClientRepository {
	return &clientRepo{db: db}
}

func (r *clientRepo) GetAll() ([]model.Client, error) {
	var rows []clientDB
	if err := r.db.Select(&rows, `SELECT id, email, name, surname, is_vip FROM clients ORDER BY id`); err != nil {
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
	if err := r.db.Get(&row, `SELECT id, email, name, surname, is_vip FROM clients WHERE id = $1`, id); err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.GetByID: %w", err)
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Create(tx *sqlx.Tx, c model.Client) (model.Client, error) {
	row := toClientDB(c)
	err := tx.QueryRowx(
		`INSERT INTO clients (email, name, surname, is_vip) VALUES ($1, $2, $3, $4) RETURNING id`,
		row.Email,
		row.Name,
		row.Surname,
		row.IsVIP,
	).Scan(&row.ID)
	if err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Update(tx *sqlx.Tx, id int, c model.Client) (model.Client, error) {
	row := toClientDB(c)
	row.ID = id
	res, err := tx.Exec(
		`UPDATE clients SET email=$1, name=$2, surname=$3, is_vip=$4 WHERE id=$5`,
		row.Email, row.Name, row.Surname, row.IsVIP, row.ID,
	)
	if err != nil {
		return model.Client{}, fmt.Errorf("clientRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Client{}, err
	}
	return row.toDomain(), nil
}

func (r *clientRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM clients WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("clientRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

func (r *clientRepo) GetEmails() ([]string, error) {
	var rows []string
	if err := r.db.Select(&rows, `SELECT email FROM clients`); err != nil {
		return nil, fmt.Errorf("clientRepo.GetEmails: %w", err)
	}
	return rows, nil
}
