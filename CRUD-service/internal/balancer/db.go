package balancer

import (
	"context"
	"time"

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

// calculateAppealPriority computes a numeric priority score for an appeal.
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

func (db *DB) RemovePendingAppeal(ctx context.Context, appealID int) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM pending_appeals WHERE appeal_id = $1`, appealID)
	return err
}

func (db *DB) FetchPendingAppeals(ctx context.Context, limit int) ([]AppealRow, error) {
	const q = `
SELECT a.id, a.team_id, a.is_urgent, a.is_important, a.created_at,
       a.pending_client_message_created_at, a.employee_id, a.status
FROM pending_appeals p
JOIN appeals a ON a.id = p.appeal_id
WHERE a.employee_id IS NULL AND a.status <> 'closed'
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
		if err := rows.Scan(
			&a.ID, &a.TeamID, &a.IsUrgent, &a.IsImportant, &a.CreatedAt,
			&a.PendingClientMessageAt, &a.ManagerID, &a.Status,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (db *DB) FetchAvailableManagers(ctx context.Context, limit int) ([]ManagerRow, error) {
	// employees — это менеджеры монолита; teams_employees — их команды.
	const q = `
SELECT e.id,
       e.is_available,
       e.active_appeals_count,
       e.last_assign_at,
       COALESCE(array_agg(te.team_id) FILTER (WHERE te.team_id IS NOT NULL), '{}') AS team_ids
FROM employees e
LEFT JOIN teams_employees te ON te.employee_id = e.id
WHERE e.is_available = true
GROUP BY e.id
ORDER BY e.active_appeals_count ASC, COALESCE(e.last_assign_at, '1970-01-01'::timestamptz) ASC
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
