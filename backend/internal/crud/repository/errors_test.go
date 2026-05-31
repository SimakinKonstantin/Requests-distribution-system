package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapNotFound(t *testing.T) {
	other := errors.New("other")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "sql no rows", err: sql.ErrNoRows, want: ErrNotFound},
		{name: "other error", err: other, want: other},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := wrapNotFound(tc.err)
			if tc.want == ErrNotFound {
				assert.True(t, errors.Is(got, ErrNotFound))
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestExpectOneRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tests := []struct {
		name      string
		sql       string
		rows      int64
		execErr   error
		wantErr   error
	}{
		{
			name:    "one row affected",
			sql:     "UPDATE clients SET name='x' WHERE id=1",
			rows:    1,
		},
		{
			name:    "zero rows affected",
			sql:     "DELETE FROM clients WHERE id=999",
			rows:    0,
			wantErr: ErrNotFound,
		},
		{
			name:    "exec fails",
			sql:     "UPDATE clients SET name='x' WHERE id=1",
			execErr: errors.New("exec failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.execErr != nil {
				mock.ExpectExec("UPDATE").WillReturnError(tc.execErr)
				_, err := db.Exec(tc.sql)
				require.Error(t, err)
				return
			}

			if tc.name == "one row affected" {
				mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(0, tc.rows))
			} else {
				mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, tc.rows))
			}
			res, err := db.Exec(tc.sql)
			require.NoError(t, err)
			if tc.wantErr != nil {
				assert.ErrorIs(t, expectOneRow(res), tc.wantErr)
				return
			}
			assert.NoError(t, expectOneRow(res))
		})
	}
}
