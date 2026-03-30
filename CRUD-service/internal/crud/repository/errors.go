package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("entity not found")

// wrapNotFound converts sql.ErrNoRows into ErrNotFound.
func wrapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// expectOneRow returns ErrNotFound if the statement affected 0 rows.
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
