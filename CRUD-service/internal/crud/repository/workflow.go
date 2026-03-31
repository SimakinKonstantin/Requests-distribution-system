package repository

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type WorkflowRepository struct {
	db *sqlx.DB
}

type WorkflowDB struct {
	ID     int             `db:"id"`
	Name   string          `db:"name"`
	Status string          `db:"status"`
	Data   json.RawMessage `db:"data"`
}

func NewWorkflowRepository(db *sqlx.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Создает workflow с такими полями, игнорирует id.
func (r *WorkflowRepository) Create(workflow model.Workflow) (int, error) {

}

type GetAllDB struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (r *WorkflowRepository) GetAll() ([]GetAllDB, error) {
	query := `SELECT id, name FROM workflows`
	var rows []GetAllDB
	if err := r.db.Select(&rows, query); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *WorkflowRepository) GetByID(id int) (WorkflowDB, error) {
	query := `SELECT id, name, status, data FROM workflows WHERE id = $1`
	var row WorkflowDB
	if err := r.db.Get(&row, query, id); err != nil {
		return WorkflowDB{}, err
	}
	return row, nil
}
