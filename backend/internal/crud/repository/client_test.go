package repository

import (
	"database/sql"
	"testing"

	"crud-service/internal/crud/model"
	"crud-service/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRepository(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "full client repository flow"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewClientRepository(db)

			mock.ExpectQuery(`SELECT id, email, name, surname, is_vip FROM clients ORDER BY id`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "surname", "is_vip"}).
					AddRow(1, "a@b.c", "A", "B", true))
			items, err := repo.GetAll()
			require.NoError(t, err)
			require.Len(t, items, 1)
			assert.True(t, items[0].IsVIP)

			mock.ExpectQuery(`SELECT id, email, name, surname, is_vip FROM clients WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "surname", "is_vip"}).
					AddRow(1, "a@b.c", "A", "B", false))
			client, err := repo.GetByID(1)
			require.NoError(t, err)
			assert.Equal(t, "a@b.c", client.Email)

			mock.ExpectQuery(`SELECT email FROM clients`).
				WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("x@y.z"))
			emails, err := repo.GetEmails()
			require.NoError(t, err)
			assert.Equal(t, []string{"x@y.z"}, emails)

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO clients`).
				WithArgs("n@e.w", "N", "E", false).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
			mock.ExpectCommit()
			tx, err := db.Beginx()
			require.NoError(t, err)
			created, err := repo.Create(tx, model.Client{Email: "n@e.w", Name: "N", Surname: "E"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			assert.Equal(t, 5, created.ID)

			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE clients SET email=\$1, name=\$2, surname=\$3, is_vip=\$4 WHERE id=\$5`).
				WithArgs("u@e.w", "U", "S", true, 5).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			updated, err := repo.Update(tx, 5, model.Client{Email: "u@e.w", Name: "U", Surname: "S", IsVIP: true})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			assert.Equal(t, 5, updated.ID)

			mock.ExpectBegin()
			mock.ExpectExec(`DELETE FROM clients WHERE id=\$1`).WithArgs(5).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			require.NoError(t, repo.Delete(tx, 5))
			require.NoError(t, tx.Commit())

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestClientRepositoryGetByIDNotFound(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "not found", id: 999},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewClientRepository(db)
			mock.ExpectQuery(`SELECT id, email, name, surname, is_vip FROM clients WHERE id = \$1`).
				WithArgs(tc.id).
				WillReturnError(sql.ErrNoRows)
			_, err := repo.GetByID(tc.id)
			assert.Error(t, err)
		})
	}
}

func TestClientDBConversion(t *testing.T) {
	tests := []struct {
		name string
		in   model.Client
	}{
		{name: "conversion roundtrip", in: model.Client{ID: 1, Email: "e", Name: "n", Surname: "s", IsVIP: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			row := toClientDB(tc.in)
			assert.Equal(t, tc.in, row.toDomain())
		})
	}
}

func expectTx(t *testing.T, db *sqlx.DB, mock sqlmock.Sqlmock) *sqlx.Tx {
	t.Helper()
	mock.ExpectBegin()
	tx, err := db.Beginx()
	require.NoError(t, err)
	return tx
}
