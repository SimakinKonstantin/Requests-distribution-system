package service

import (
	"database/sql"
	"errors"
	"testing"

	"crud-service/internal/crud/model"
	"crud-service/internal/testutil"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemeServiceCRUD(t *testing.T) {
	tests := []struct{ name string }{{name: "theme service crud"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := &mockThemeRepo{
				getAllFn: func() ([]model.Theme, error) { return []model.Theme{{ID: 1}}, nil },
				getByIDFn: func(id int) (model.Theme, error) { return model.Theme{ID: id}, nil },
				createFn: func(tx *sqlx.Tx, t model.Theme) (model.Theme, error) { t.ID = 2; return t, nil },
				updateFn: func(tx *sqlx.Tx, id int, t model.Theme) (model.Theme, error) { t.ID = id; return t, nil },
				deleteFn: func(tx *sqlx.Tx, id int) error { return nil },
			}
			svc := NewThemeService(db, repo)

			items, err := svc.GetAll()
			require.NoError(t, err)
			assert.Len(t, items, 1)

			item, err := svc.GetByID(1)
			require.NoError(t, err)
			assert.Equal(t, 1, item.ID)

			mock.ExpectBegin()
			mock.ExpectCommit()
			created, err := svc.Create(model.Theme{Name: "T"})
			require.NoError(t, err)
			assert.Equal(t, 2, created.ID)

			mock.ExpectBegin()
			mock.ExpectCommit()
			updated, err := svc.Update(2, model.Theme{Name: "U"})
			require.NoError(t, err)
			assert.Equal(t, 2, updated.ID)

			mock.ExpectBegin()
			mock.ExpectCommit()
			require.NoError(t, svc.Delete(2))
		})
	}
}

func TestSubthemeServiceCRUD(t *testing.T) {
	tests := []struct{ name string }{{name: "subtheme service crud"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := &mockSubthemeRepo{
				getAllFn: func() ([]model.Subtheme, error) { return nil, nil },
				createFn: func(tx *sqlx.Tx, s model.Subtheme) (model.Subtheme, error) { s.ID = 1; return s, nil },
			}
			svc := NewSubthemeService(db, repo)

			mock.ExpectBegin()
			mock.ExpectCommit()
			created, err := svc.Create(model.Subtheme{Name: "S"})
			require.NoError(t, err)
			assert.Equal(t, 1, created.ID)
		})
	}
}

func TestTeamServiceAssignTeam(t *testing.T) {
	tests := []struct {
		name     string
		appealID int
		teamID   int
	}{{name: "assign team", appealID: 1, teamID: 2}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := &mockTeamRepo{
				assignFn: func(tx *sqlx.Tx, appealID, teamID int) error {
					assert.Equal(t, tc.appealID, appealID)
					assert.Equal(t, tc.teamID, teamID)
					return nil
				},
			}
			svc := NewTeamService(db, repo)

			mock.ExpectBegin()
			mock.ExpectCommit()
			require.NoError(t, svc.AssignTeam(tc.appealID, tc.teamID))
		})
	}
}

func TestClientServiceUpdateDelete(t *testing.T) {
	tests := []struct{ name string }{{name: "update and delete client"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := &mockClientRepo{
				updateFn: func(tx *sqlx.Tx, id int, c model.Client) (model.Client, error) { c.ID = id; return c, nil },
				deleteFn: func(tx *sqlx.Tx, id int) error { return nil },
			}
			svc := NewClientService(db, repo)

			mock.ExpectBegin()
			mock.ExpectCommit()
			updated, err := svc.Update(3, model.Client{Email: "x@y.z"})
			require.NoError(t, err)
			assert.Equal(t, 3, updated.ID)

			mock.ExpectBegin()
			mock.ExpectCommit()
			require.NoError(t, svc.Delete(3))
		})
	}
}

func TestAppealServiceFetchPendingAppeals(t *testing.T) {
	tests := []struct{ name string }{{name: "fetch pending appeals"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &appealService{
				appealRepo: &mockAppealRepoWithFetch{
					mockAppealRepo: mockAppealRepo{getByIDFn: func(id int) (model.Appeal, error) {
						return model.Appeal{ID: id, ClientID: 1}, nil
					}},
					fetchFn: func(limit int) ([]model.Appeal, error) {
						return []model.Appeal{{ID: 1, ClientID: 1}}, nil
					},
				},
				clientRepo: &mockAppealClientRepo{
					getByIDFn: func(id int) (model.Client, error) { return model.Client{IsVIP: false}, nil },
				},
			}
			items, err := svc.FetchPendingAppeals(10)
			require.NoError(t, err)
			require.Len(t, items, 1)
			assert.False(t, items[0].IsImportant)
		})
	}
}

type mockAppealRepoWithFetch struct {
	mockAppealRepo
	fetchFn func(limit int) ([]model.Appeal, error)
}

func (m *mockAppealRepoWithFetch) FetchPendingAppeals(limit int) ([]model.Appeal, error) {
	if m.fetchFn != nil {
		return m.fetchFn(limit)
	}
	return nil, nil
}

func TestAppealServiceUpsertPendingAssigned(t *testing.T) {
	tests := []struct{ name string }{{name: "assigned appeal removed from pending"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removed := false
			empID := 5
			svc := &appealService{
				appealRepo: &mockAppealRepo{
					getByIDFn: func(id int) (model.Appeal, error) {
						return model.Appeal{ID: id, Status: "active", EmployeeID: &empID}, nil
					},
				},
				pendingAppealRepo: &mockPendingAppealRepo{
					removeFn: func(appealID int) error { removed = true; return nil },
				},
			}
			require.NoError(t, svc.UpsertPendingAppealByID(1))
			assert.True(t, removed)
		})
	}
}

func TestAppealServiceUpsertPendingUpsert(t *testing.T) {
	tests := []struct{ name string }{{name: "upsert pending appeal"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upserted := false
			svc := &appealService{
				appealRepo: &mockAppealRepo{
					getByIDFn: func(id int) (model.Appeal, error) {
						return model.Appeal{ID: id, Status: "active"}, nil
					},
				},
				clientRepo: &mockAppealClientRepo{
					getByIDFn: func(id int) (model.Client, error) { return model.Client{}, nil },
				},
				pendingAppealRepo: &mockPendingAppealRepo{
					upsertFn: func(appealID int, priority int64) error {
						upserted = true
						assert.Equal(t, 1, appealID)
						assert.Greater(t, priority, int64(0))
						return nil
					},
				},
			}
			require.NoError(t, svc.UpsertPendingAppealByID(1))
			assert.True(t, upserted)
		})
	}
}

func TestAppealServiceGetAll(t *testing.T) {
	tests := []struct{ name string }{{name: "get all appeals"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &appealService{
				appealRepo: &mockAppealRepo{
					getAllFn: func() ([]model.Appeal, error) { return []model.Appeal{{ID: 1}}, nil },
				},
			}
			items, err := svc.GetAll()
			require.NoError(t, err)
			assert.Len(t, items, 1)
		})
	}
}

func TestSlotServiceUpdateCountSame(t *testing.T) {
	tests := []struct{ name string }{{name: "same slot count"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			svc := NewSlotService(db, &mockSlotRepo{
				getCountFn: func(employeeID int) (int, error) { return 3, nil },
			})

			mock.ExpectBegin()
			mock.ExpectCommit()
			require.NoError(t, svc.UpdateCount(1, 3))
		})
	}
}

type mockThemeRepo struct {
	getAllFn  func() ([]model.Theme, error)
	getByIDFn func(id int) (model.Theme, error)
	createFn  func(tx *sqlx.Tx, t model.Theme) (model.Theme, error)
	updateFn  func(tx *sqlx.Tx, id int, t model.Theme) (model.Theme, error)
	deleteFn  func(tx *sqlx.Tx, id int) error
}

func (m *mockThemeRepo) GetAll() ([]model.Theme, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockThemeRepo) GetByID(id int) (model.Theme, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Theme{}, nil
}
func (m *mockThemeRepo) Create(tx *sqlx.Tx, t model.Theme) (model.Theme, error) {
	if m.createFn != nil {
		return m.createFn(tx, t)
	}
	return t, nil
}
func (m *mockThemeRepo) Update(tx *sqlx.Tx, id int, t model.Theme) (model.Theme, error) {
	if m.updateFn != nil {
		return m.updateFn(tx, id, t)
	}
	return t, nil
}
func (m *mockThemeRepo) Delete(tx *sqlx.Tx, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(tx, id)
	}
	return nil
}

type mockSubthemeRepo struct {
	getAllFn  func() ([]model.Subtheme, error)
	getByIDFn func(id int) (model.Subtheme, error)
	createFn  func(tx *sqlx.Tx, s model.Subtheme) (model.Subtheme, error)
	updateFn  func(tx *sqlx.Tx, id int, s model.Subtheme) (model.Subtheme, error)
	deleteFn  func(tx *sqlx.Tx, id int) error
}

func (m *mockSubthemeRepo) GetAll() ([]model.Subtheme, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockSubthemeRepo) GetByID(id int) (model.Subtheme, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Subtheme{}, nil
}
func (m *mockSubthemeRepo) Create(tx *sqlx.Tx, s model.Subtheme) (model.Subtheme, error) {
	if m.createFn != nil {
		return m.createFn(tx, s)
	}
	return s, nil
}
func (m *mockSubthemeRepo) Update(tx *sqlx.Tx, id int, s model.Subtheme) (model.Subtheme, error) {
	if m.updateFn != nil {
		return m.updateFn(tx, id, s)
	}
	return s, nil
}
func (m *mockSubthemeRepo) Delete(tx *sqlx.Tx, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(tx, id)
	}
	return nil
}

func TestAppealServiceGetByIDError(t *testing.T) {
	tests := []struct{ name string }{{name: "get by id error"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &appealService{
				appealRepo: &mockAppealRepo{getByIDFn: func(int) (model.Appeal, error) { return model.Appeal{}, errors.New("fail") }},
			}
			_, err := svc.GetByID(1)
			assert.Error(t, err)
		})
	}
}

func TestAppealServiceCloseNoSlot(t *testing.T) {
	tests := []struct{ name string }{{name: "close without slot"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			svc := &appealService{
				db: db,
				appealRepo: &mockAppealRepo{},
				slotRepo: &mockSlotRepo{
					getByAppealFn: func(appealID int) (model.Slot, error) { return model.Slot{}, sql.ErrNoRows },
				},
			}
			svc.appealRepo = &mockAppealRepoWithClose{
				closeFn: func(tx *sqlx.Tx, id int) (model.Appeal, error) { return model.Appeal{ID: id, Status: "closed"}, nil },
			}

			mock.ExpectBegin()
			mock.ExpectCommit()
			closed, err := svc.Close(1)
			require.NoError(t, err)
			assert.Equal(t, "closed", closed.Status)
		})
	}
}

type mockAppealRepoWithClose struct {
	mockAppealRepo
	closeFn func(tx *sqlx.Tx, id int) (model.Appeal, error)
}

func (m *mockAppealRepoWithClose) Close(tx *sqlx.Tx, id int) (model.Appeal, error) {
	if m.closeFn != nil {
		return m.closeFn(tx, id)
	}
	return model.Appeal{}, nil
}
