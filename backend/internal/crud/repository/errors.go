package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("entity not found")

func wrapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func expectOneRow(res interface {
	RowsAffected() (int64, error)
}) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("RowsAffected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
