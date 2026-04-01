package workflow

import (
	"encoding/json"
	"fmt"

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

func (r *WorkflowRepository) Save(workflow WorkflowDB) (int, error) {
	err := r.db.QueryRowx(
		`INSERT INTO workflows (name, status, data) VALUES ($1, $2, $3) RETURNING id`,
		workflow.Name,
		workflow.Status,
		workflow.Data,
	).Scan(&workflow.ID)
	if err != nil {
		return -1, fmt.Errorf("failed to save workflow: %w", err)
	}
	return workflow.ID, nil
}

func (r *WorkflowRepository) All() ([]WorkflowDB, error) {
	query := `SELECT id, name, status, data FROM workflows`
	var rows []WorkflowDB
	if err := r.db.Select(&rows, query); err != nil {
		return nil, fmt.Errorf("failed to get all workflows: %w", err)
	}
	return rows, nil
}

func (r *WorkflowRepository) Update(id int, workflow WorkflowDB) error {
	_, err := r.db.Exec(
		`UPDATE workflows SET name = $1, status = $2, data = $3 WHERE id = $4`,
		workflow.Name,
		workflow.Status,
		workflow.Data,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) UpdateStatus(id int, state Status) error {
	_, err := r.db.Exec(
		`UPDATE workflows SET status = $1 WHERE id = $2`,
		state,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) Get(id int) (WorkflowDB, error) {
	query := `SELECT id, name, status, data FROM workflows WHERE id = $1`
	var row WorkflowDB
	if err := r.db.Get(&row, query, id); err != nil {
		return WorkflowDB{}, fmt.Errorf("failed to get workflow: %w", err)
	}
	return row, nil
}
