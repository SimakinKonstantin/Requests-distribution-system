package repository

import (
	"testing"

	"crud-service/internal/crud/model"
	"crud-service/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemeRepositoryCRUD(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "theme repository crud"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewThemeRepository(db)

			mock.ExpectQuery(`SELECT id, name FROM themes ORDER BY id`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Support"))
			themes, err := repo.GetAll()
			require.NoError(t, err)
			require.Len(t, themes, 1)

			mock.ExpectQuery(`SELECT id, name FROM themes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Support"))
			theme, err := repo.GetByID(1)
			require.NoError(t, err)
			assert.Equal(t, "Support", theme.Name)

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO themes`).WithArgs("New").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			mock.ExpectCommit()
			tx, err := db.Beginx()
			require.NoError(t, err)
			created, err := repo.Create(tx, model.Theme{Name: "New"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			assert.Equal(t, 2, created.ID)

			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE themes SET name=\$1 WHERE id=\$2`).WithArgs("Upd", 2).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			updated, err := repo.Update(tx, 2, model.Theme{Name: "Upd"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			assert.Equal(t, "Upd", updated.Name)

			mock.ExpectBegin()
			mock.ExpectExec(`DELETE FROM themes WHERE id=\$1`).WithArgs(2).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			require.NoError(t, repo.Delete(tx, 2))
			require.NoError(t, tx.Commit())

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSubthemeRepositoryCRUD(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "subtheme repository crud"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewSubthemeRepository(db)

			mock.ExpectQuery(`SELECT id, name FROM subthemes`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Billing"))
			items, err := repo.GetAll()
			require.NoError(t, err)
			require.Len(t, items, 1)

			mock.ExpectQuery(`SELECT id, name FROM subthemes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Billing"))
			item, err := repo.GetByID(1)
			require.NoError(t, err)
			assert.Equal(t, "Billing", item.Name)

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO subthemes`).WithArgs("Refund").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			mock.ExpectCommit()
			tx, err := db.Beginx()
			require.NoError(t, err)
			created, err := repo.Create(tx, model.Subtheme{Name: "Refund"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			assert.Equal(t, 2, created.ID)

			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE subthemes SET name=\$1 WHERE id=\$2`).WithArgs("Updated", 2).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			_, err = repo.Update(tx, 2, model.Subtheme{Name: "Updated"})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())

			mock.ExpectBegin()
			mock.ExpectExec(`DELETE FROM subthemes WHERE id=\$1`).WithArgs(2).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err = db.Beginx()
			require.NoError(t, err)
			require.NoError(t, repo.Delete(tx, 2))
			require.NoError(t, tx.Commit())
		})
	}
}

func TestTeamRepositoryGetByName(t *testing.T) {
	tests := []struct {
		name string
		team string
	}{
		{name: "find by name", team: "VIP"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewTeamRepository(db)

			mock.ExpectQuery(`SELECT id, name FROM teams WHERE name = \$1`).WithArgs(tc.team).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, tc.team))

			team, err := repo.GetTeamByName(tc.team)
			require.NoError(t, err)
			assert.Equal(t, 5, team.ID)
		})
	}
}

func TestTeamRepositoryAssignTeam(t *testing.T) {
	tests := []struct {
		name     string
		appealID int
		teamID   int
	}{
		{name: "assign team", appealID: 10, teamID: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			repo := NewTeamRepository(db)

			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE appeals SET team_id = \$1 WHERE id = \$2`).WithArgs(tc.teamID, tc.appealID).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			tx, err := db.Beginx()
			require.NoError(t, err)
			require.NoError(t, repo.AssignTeam(tx, tc.appealID, tc.teamID))
			require.NoError(t, tx.Commit())
		})
	}
}

func TestThemeSubthemeConversion(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "theme and subtheme conversions"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			theme := model.Theme{ID: 1, Name: "T"}
			assert.Equal(t, theme, toThemeDB(theme).toDomain())

			sub := model.Subtheme{ID: 2, Name: "S"}
			assert.Equal(t, sub, toSubthemeDB(sub).toDomain())
		})
	}
}
