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
