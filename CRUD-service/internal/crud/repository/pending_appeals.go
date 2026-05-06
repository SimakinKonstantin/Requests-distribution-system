package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PendingAppealDB struct {
	AppealID  int       `db:"appeal_id"`
	Priority  int       `db:"priority"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PendingAppealRepository interface {
	GetAll() ([]PendingAppealDB, error)
	GetByAppealID(appealID int) (PendingAppealDB, error)
	Create(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error
	Update(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error
	Delete(tx *sqlx.Tx, appealID int) error
	RemovePendingAppeal(appealID int) error
	UpsertPendingAppealByID(appealID int, priority int64) error
}

type pendingAppealRepo struct {
	db *sqlx.DB
}

func NewPendingAppealRepository(db *sqlx.DB) PendingAppealRepository {
	return &pendingAppealRepo{db: db}
}

func (r *pendingAppealRepo) GetAll() ([]PendingAppealDB, error) {
	var rows []PendingAppealDB
	err := r.db.Select(&rows, `SELECT appeal_id, priority, updated_at FROM pending_appeals`)
	if err != nil {
		return nil, fmt.Errorf("pendingAppealRepo.GetAll: %w", err)
	}
	return rows, nil
}

func (r *pendingAppealRepo) GetByAppealID(appealID int) (PendingAppealDB, error) {
	var row PendingAppealDB
	err := r.db.Get(&row, `SELECT appeal_id, priority, updated_at FROM pending_appeals WHERE appeal_id = $1`, appealID)
	if err != nil {
		return PendingAppealDB{}, fmt.Errorf("pendingAppealRepo.GetByAppealID: %w", err)
	}
	return row, nil
}

func (r *pendingAppealRepo) Create(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error {
	_, err := tx.Exec(`INSERT INTO pending_appeals (appeal_id, priority, updated_at) VALUES ($1, $2, $3)`, pendingAppeal.AppealID, pendingAppeal.Priority, pendingAppeal.UpdatedAt)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.Create: %w", err)
	}
	return nil
}

func (r *pendingAppealRepo) Update(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error {
	_, err := tx.Exec(`UPDATE pending_appeals SET priority = $1, updated_at = $2 WHERE appeal_id = $3`, pendingAppeal.Priority, pendingAppeal.UpdatedAt, pendingAppeal.AppealID)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.Update: %w", err)
	}
	return nil
}

func (r *pendingAppealRepo) Delete(tx *sqlx.Tx, appealID int) error {
	_, err := tx.Exec(`DELETE FROM pending_appeals WHERE appeal_id = $1`, appealID)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.Delete: %w", err)
	}
	return nil
}

func (r *pendingAppealRepo) RemovePendingAppeal(appealID int) error {
	_, err := r.db.Exec(`DELETE FROM pending_appeals WHERE appeal_id = $1`, appealID)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.RemovePendingAppeal: %w", err)
	}
	return nil
}

func (r *pendingAppealRepo) UpsertPendingAppealByID(appealID int, priority int64) error {
	query := `
INSERT INTO pending_appeals (appeal_id, priority)
VALUES ($1, $2)
ON CONFLICT (appeal_id) DO UPDATE SET
  priority=excluded.priority,
  updated_at=now()`

	_, err := r.db.Exec(query, appealID, priority)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.UpsertPendingAppealByID: %w", err)
	}
	return nil
}
