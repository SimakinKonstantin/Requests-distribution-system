package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PendingAppealDB struct {
	AppealID  int       `db:"appeal_id"`
	TeamID    int       `db:"team_id"`
	Priority  int       `db:"priority"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PendingAppealRepository interface {
	GetAll() ([]PendingAppealDB, error)
	GetByAppealID(appealID int) (PendingAppealDB, error)
	Create(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error
	Update(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error
	Delete(tx *sqlx.Tx, appealID int) error
}

type pendingAppealRepo struct {
	db *sqlx.DB
}

func NewPendingAppealRepository(db *sqlx.DB) PendingAppealRepository {
	return &pendingAppealRepo{db: db}
}

func (r *pendingAppealRepo) GetAll() ([]PendingAppealDB, error) {
	var rows []PendingAppealDB
	err := r.db.Select(&rows, `SELECT appeal_id, team_id, priority, updated_at FROM pending_appeals`)
	if err != nil {
		return nil, fmt.Errorf("pendingAppealRepo.GetAll: %w", err)
	}
	return rows, nil
}

func (r *pendingAppealRepo) GetByAppealID(appealID int) (PendingAppealDB, error) {
	var row PendingAppealDB
	err := r.db.Get(&row, `SELECT appeal_id, team_id, priority, updated_at FROM pending_appeals WHERE appeal_id = $1`, appealID)
	if err != nil {
		return PendingAppealDB{}, fmt.Errorf("pendingAppealRepo.GetByAppealID: %w", err)
	}
	return row, nil
}

func (r *pendingAppealRepo) Create(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error {
	_, err := tx.Exec(`INSERT INTO pending_appeals (appeal_id, team_id, priority, updated_at) VALUES ($1, $2, $3, $4)`, pendingAppeal.AppealID, pendingAppeal.TeamID, pendingAppeal.Priority, pendingAppeal.UpdatedAt)
	if err != nil {
		return fmt.Errorf("pendingAppealRepo.Create: %w", err)
	}
	return nil
}

func (r *pendingAppealRepo) Update(tx *sqlx.Tx, pendingAppeal PendingAppealDB) error {
	_, err := tx.Exec(`UPDATE pending_appeals SET team_id = $1, priority = $2, updated_at = $3 WHERE appeal_id = $4`, pendingAppeal.TeamID, pendingAppeal.Priority, pendingAppeal.UpdatedAt, pendingAppeal.AppealID)
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
