package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() { db.Pool.Close() }

// calculateAppealPriority ported from src/services/state/state.service.ts
func calculateAppealPriority(isImportant, isUrgent bool, createdAt time.Time, pendingClientMsgAt *time.Time, now time.Time) float64 {
	group := 3
	switch {
	case isImportant && isUrgent:
		group = 0
	case isImportant && !isUrgent:
		group = 1
	case !isImportant && isUrgent:
		group = 2
	default:
		group = 3
	}

	priority := float64((3 - group) * 1_000_000)

	if pendingClientMsgAt != nil {
		ageMinutes := now.Sub(*pendingClientMsgAt).Minutes()
		priority += ageMinutes * 10
	}

	priority += now.Sub(createdAt).Minutes()
	return priority
}

func (db *DB) UpsertPendingAppealByID(ctx context.Context, appealID int) error {
	// For demo: priority is computed from appeals table fields.
	const q = `
SELECT id, team_id, is_urgent, is_important, created_at, pending_client_message_created_at, manager_id, status
FROM appeals
WHERE id=$1
`
	var a AppealRow
	err := db.Pool.QueryRow(ctx, q, appealID).Scan(
		&a.ID, &a.TeamID, &a.IsUrgent, &a.IsImportant, &a.CreatedAt, &a.PendingClientMessageAt, &a.ManagerID, &a.Status,
	)
	if err != nil {
		return err
	}
	if a.Status == "closed" {
		return db.RemovePendingAppeal(ctx, appealID)
	}
	if a.ManagerID != nil {
		// already assigned → should not be pending
		return db.RemovePendingAppeal(ctx, appealID)
	}
	now := time.Now().UTC()
	priority := calculateAppealPriority(a.IsImportant, a.IsUrgent, a.CreatedAt, a.PendingClientMessageAt, now)

	const upsert = `
INSERT INTO pending_appeals (appeal_id, team_id, priority, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (appeal_id) DO UPDATE SET
  team_id=excluded.team_id,
  priority=excluded.priority,
  updated_at=now()
`
	_, err = db.Pool.Exec(ctx, upsert, a.ID, a.TeamID, priority)
	return err
}

func (db *DB) RemovePendingAppeal(ctx context.Context, appealID int) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM pending_appeals WHERE appeal_id=$1`, appealID)
	return err
}

func (db *DB) FetchPendingAppeals(ctx context.Context, limit int) ([]AppealRow, error) {
	const q = `
SELECT a.id, a.team_id, a.is_urgent, a.is_important, a.created_at, a.pending_client_message_created_at, a.manager_id, a.status
FROM pending_appeals p
JOIN appeals a ON a.id=p.appeal_id
WHERE a.manager_id IS NULL AND a.status <> 'closed'
ORDER BY p.priority DESC, p.updated_at ASC
LIMIT $1
`
	rows, err := db.Pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppealRow
	for rows.Next() {
		var a AppealRow
		if err := rows.Scan(&a.ID, &a.TeamID, &a.IsUrgent, &a.IsImportant, &a.CreatedAt, &a.PendingClientMessageAt, &a.ManagerID, &a.Status); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (db *DB) FetchAvailableManagers(ctx context.Context, limit int) ([]ManagerRow, error) {
	// For demo: manager_teams gives team membership, slots with appeal_id NULL are free.
	const q = `
SELECT m.id,
       m.is_available,
       m.active_appeals_count,
       m.last_assign_at,
       COALESCE(array_agg(mt.team_id) FILTER (WHERE mt.team_id IS NOT NULL), '{}') AS team_ids
FROM managers m
LEFT JOIN manager_teams mt ON mt.manager_id=m.id
WHERE m.is_available = true
GROUP BY m.id
ORDER BY m.active_appeals_count ASC, COALESCE(m.last_assign_at, '1970-01-01'::timestamptz) ASC
LIMIT $1
`
	rows, err := db.Pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ManagerRow
	for rows.Next() {
		var m ManagerRow
		if err := rows.Scan(&m.ID, &m.IsAvailable, &m.ActiveAppeals, &m.LastAssignAt, &m.TeamIDs); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (db *DB) FetchFreeSlotsByManagers(ctx context.Context, managerIDs []string) (map[string][]SlotRow, error) {
	if len(managerIDs) == 0 {
		return map[string][]SlotRow{}, nil
	}
	const q = `
SELECT id, manager_id, appeal_id, updated_at
FROM slots
WHERE manager_id = ANY($1) AND appeal_id IS NULL
ORDER BY manager_id, updated_at ASC
`
	rows, err := db.Pool.Query(ctx, q, managerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]SlotRow, len(managerIDs))
	for rows.Next() {
		var s SlotRow
		if err := rows.Scan(&s.ID, &s.ManagerID, &s.AppealID, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out[s.ManagerID] = append(out[s.ManagerID], s)
	}
	return out, rows.Err()
}

func (db *DB) CloseAppeal(ctx context.Context, appealID int) error {
	// Simplified: mark closed, free slot, drop from pending
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var managerID *string
	err = tx.QueryRow(ctx, `SELECT manager_id FROM appeals WHERE id=$1 FOR UPDATE`, appealID).Scan(&managerID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE appeals SET status='closed' WHERE id=$1`, appealID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM pending_appeals WHERE appeal_id=$1`, appealID)
	if err != nil {
		return err
	}
	if managerID != nil {
		_, err = tx.Exec(ctx, `UPDATE slots SET appeal_id=NULL, updated_at=now() WHERE appeal_id=$1`, appealID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE managers SET active_appeals_count = GREATEST(active_appeals_count-1, 0) WHERE id=$1`, *managerID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

var ErrNoFreeSlot = errors.New("no free slot")
var ErrAppealAlreadyAssigned = errors.New("appeal already assigned")

func (db *DB) AssignAppealTx(ctx context.Context, appealID int, managerID string, slotID string) error {
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingManagerID *string
	var status string
	if err := tx.QueryRow(ctx, `SELECT manager_id, status FROM appeals WHERE id=$1 FOR UPDATE`, appealID).Scan(&existingManagerID, &status); err != nil {
		return err
	}
	if status == "closed" {
		return fmt.Errorf("appeal %d is closed", appealID)
	}
	if existingManagerID != nil {
		return ErrAppealAlreadyAssigned
	}

	// Ensure slot is free and belongs to manager.
	var slotAppealID *int
	if err := tx.QueryRow(ctx, `SELECT appeal_id FROM slots WHERE id=$1 AND manager_id=$2 FOR UPDATE`, slotID, managerID).Scan(&slotAppealID); err != nil {
		return err
	}
	if slotAppealID != nil {
		return ErrNoFreeSlot
	}

	_, err = tx.Exec(ctx, `UPDATE appeals SET manager_id=$1, status='assigned' WHERE id=$2`, managerID, appealID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE slots SET appeal_id=$1, updated_at=now() WHERE id=$2`, appealID, slotID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM pending_appeals WHERE appeal_id=$1`, appealID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE managers SET active_appeals_count=active_appeals_count+1, last_assign_at=now() WHERE id=$1`, managerID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

