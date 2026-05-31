package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"crud-service/internal/testutil"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClientRepo struct {
	getAllFn   func() ([]model.Client, error)
	getByIDFn  func(id int) (model.Client, error)
	createFn   func(tx *sqlx.Tx, c model.Client) (model.Client, error)
	updateFn   func(tx *sqlx.Tx, id int, c model.Client) (model.Client, error)
	deleteFn   func(tx *sqlx.Tx, id int) error
	getEmailsFn func() ([]string, error)
}

func (m *mockClientRepo) GetAll() ([]model.Client, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockClientRepo) GetByID(id int) (model.Client, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Client{}, nil
}
func (m *mockClientRepo) Create(tx *sqlx.Tx, c model.Client) (model.Client, error) {
	if m.createFn != nil {
		return m.createFn(tx, c)
	}
	return c, nil
}
func (m *mockClientRepo) Update(tx *sqlx.Tx, id int, c model.Client) (model.Client, error) {
	if m.updateFn != nil {
		return m.updateFn(tx, id, c)
	}
	return c, nil
}
func (m *mockClientRepo) Delete(tx *sqlx.Tx, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(tx, id)
	}
	return nil
}
func (m *mockClientRepo) GetEmails() ([]string, error) {
	if m.getEmailsFn != nil {
		return m.getEmailsFn()
	}
	return nil, nil
}

func TestClientServiceCreate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "create client success"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := &mockClientRepo{
				createFn: func(tx *sqlx.Tx, c model.Client) (model.Client, error) {
					c.ID = 1
					return c, nil
				},
			}
			svc := NewClientService(db, repo)

			mock.ExpectBegin()
			mock.ExpectCommit()
			created, err := svc.Create(model.Client{Email: "a@b.c"})
			require.NoError(t, err)
			assert.Equal(t, 1, created.ID)
		})
	}
}

func TestClientServiceCreateBeginError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "begin error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()
			mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
			svc := NewClientService(db, &mockClientRepo{})
			_, err := svc.Create(model.Client{})
			assert.Error(t, err)
		})
	}
}

func TestClientServiceGetAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get all clients"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := testutil.NewMockDB(t)
			defer db.Close()
			svc := NewClientService(db, &mockClientRepo{
				getAllFn: func() ([]model.Client, error) { return []model.Client{{ID: 1}}, nil },
			})
			items, err := svc.GetAll()
			require.NoError(t, err)
			assert.Len(t, items, 1)
		})
	}
}

func TestCalculateAppealPriority(t *testing.T) {
	tests := []struct {
		name      string
		important bool
	}{
		{name: "regular", important: false},
		{name: "vip", important: true},
	}
	created := time.Now().Add(-10 * time.Minute)
	priorityRegular := int64(0)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			priority := calculateAppealPriority(tc.important, created)
			assert.Greater(t, priority, int64(100))
			if !tc.important {
				priorityRegular = priority
			} else {
				assert.Greater(t, priority, priorityRegular)
			}
		})
	}
}

func TestTeamServiceGetTeamFallback(t *testing.T) {
	tests := []struct {
		name   string
		isVIP  bool
		wantID int
		wantName string
	}{
		{name: "fallback by theme", isVIP: false, wantID: 10},
		{name: "fallback by default vip", isVIP: true, wantName: DEFAULT_TEAM_NAME_VIP},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			subthemeID := 2
			repo := &mockTeamRepo{
				getByThemeFn: func(themeID int, sub *int, vip bool) (model.Team, error) {
					if sub != nil {
						return model.Team{}, sql.ErrNoRows
					}
					if themeID == 1 && !vip {
						return model.Team{ID: 10, Name: "team"}, nil
					}
					return model.Team{}, sql.ErrNoRows
				},
				getByNameFn: func(name string) (model.Team, error) {
					return model.Team{ID: 99, Name: name}, nil
				},
			}
			svc := NewTeamService(db, repo)
			team, err := svc.GetTeam(1, &subthemeID, tc.isVIP)
			require.NoError(t, err)
			if tc.wantID != 0 {
				assert.Equal(t, tc.wantID, team.ID)
			}
			if tc.wantName != "" {
				assert.Equal(t, tc.wantName, team.Name)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

type mockTeamRepo struct {
	getAllFn     func() ([]model.Team, error)
	getByIDFn    func(id int) (model.Team, error)
	createFn     func(tx *sqlx.Tx, t model.Team) (model.Team, error)
	updateFn     func(tx *sqlx.Tx, id int, t model.Team) (model.Team, error)
	deleteFn     func(tx *sqlx.Tx, id int) error
	assignFn     func(tx *sqlx.Tx, appealID, teamID int) error
	getByThemeFn func(themeID int, subthemeID *int, isVIP bool) (model.Team, error)
	getByNameFn  func(name string) (model.Team, error)
}

func (m *mockTeamRepo) GetAll() ([]model.Team, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockTeamRepo) GetByID(id int) (model.Team, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Team{}, nil
}
func (m *mockTeamRepo) Create(tx *sqlx.Tx, t model.Team) (model.Team, error) {
	if m.createFn != nil {
		return m.createFn(tx, t)
	}
	return t, nil
}
func (m *mockTeamRepo) Update(tx *sqlx.Tx, id int, t model.Team) (model.Team, error) {
	if m.updateFn != nil {
		return m.updateFn(tx, id, t)
	}
	return t, nil
}
func (m *mockTeamRepo) Delete(tx *sqlx.Tx, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(tx, id)
	}
	return nil
}
func (m *mockTeamRepo) AssignTeam(tx *sqlx.Tx, appealID, teamID int) error {
	if m.assignFn != nil {
		return m.assignFn(tx, appealID, teamID)
	}
	return nil
}
func (m *mockTeamRepo) GetTeamByThemeSubtheme(themeID int, subthemeID *int, isVIP bool) (model.Team, error) {
	if m.getByThemeFn != nil {
		return m.getByThemeFn(themeID, subthemeID, isVIP)
	}
	return model.Team{}, nil
}
func (m *mockTeamRepo) GetTeamByName(name string) (model.Team, error) {
	if m.getByNameFn != nil {
		return m.getByNameFn(name)
	}
	return model.Team{}, nil
}

func TestAppealServiceIsImportant(t *testing.T) {
	tests := []struct {
		name string
		vip  bool
		want bool
	}{
		{name: "vip client", vip: true, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &appealService{
				appealRepo: &mockAppealRepo{
					getByIDFn: func(id int) (model.Appeal, error) { return model.Appeal{ClientID: 1}, nil },
				},
				clientRepo: &mockAppealClientRepo{
					getByIDFn: func(id int) (model.Client, error) { return model.Client{IsVIP: tc.vip}, nil },
				},
			}
			assert.Equal(t, tc.want, svc.IsImportant(1))
		})
	}
}

func TestAppealServiceIsImportantErrors(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "appeal repo error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &appealService{
				appealRepo: &mockAppealRepo{getByIDFn: func(int) (model.Appeal, error) { return model.Appeal{}, errors.New("fail") }},
			}
			assert.False(t, svc.IsImportant(1))
		})
	}
}

type mockAppealRepo struct {
	getAllFn func() ([]model.Appeal, error)
	getByIDFn func(id int) (model.Appeal, error)
}

func (m *mockAppealRepo) GetAll() ([]model.Appeal, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockAppealRepo) GetByID(id int) (model.Appeal, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Appeal{}, nil
}
func (m *mockAppealRepo) Create(tx *sqlx.Tx, a model.Appeal) (model.Appeal, error) { return a, nil }
func (m *mockAppealRepo) Update(tx *sqlx.Tx, id int, a model.Appeal) (model.Appeal, error) {
	return a, nil
}
func (m *mockAppealRepo) Delete(tx *sqlx.Tx, id int) error { return nil }
func (m *mockAppealRepo) Close(tx *sqlx.Tx, id int) (model.Appeal, error) {
	return model.Appeal{ID: id}, nil
}
func (m *mockAppealRepo) FetchPendingAppeals(limit int) ([]model.Appeal, error) { return nil, nil }
func (m *mockAppealRepo) AssignTeam(tx *sqlx.Tx, appealID, teamID int) error { return nil }

type mockAppealClientRepo struct {
	getByIDFn func(id int) (model.Client, error)
}

func (m *mockAppealClientRepo) GetAll() ([]model.Client, error) { return nil, nil }
func (m *mockAppealClientRepo) GetByID(id int) (model.Client, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Client{}, nil
}
func (m *mockAppealClientRepo) Create(tx *sqlx.Tx, c model.Client) (model.Client, error) { return c, nil }
func (m *mockAppealClientRepo) Update(tx *sqlx.Tx, id int, c model.Client) (model.Client, error) {
	return c, nil
}
func (m *mockAppealClientRepo) Delete(tx *sqlx.Tx, id int) error { return nil }
func (m *mockAppealClientRepo) GetEmails() ([]string, error) { return nil, nil }

func TestAppealServiceUpsertPendingAppealClosed(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "closed appeal is removed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removed := false
			svc := &appealService{
				appealRepo: &mockAppealRepo{
					getByIDFn: func(id int) (model.Appeal, error) {
						return model.Appeal{ID: id, Status: "closed"}, nil
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

type mockPendingAppealRepo struct {
	upsertFn func(appealID int, priority int64) error
	removeFn func(appealID int) error
}

func (m *mockPendingAppealRepo) GetAll() ([]repository.PendingAppealDB, error) { return nil, nil }
func (m *mockPendingAppealRepo) GetByAppealID(appealID int) (repository.PendingAppealDB, error) {
	return repository.PendingAppealDB{}, nil
}
func (m *mockPendingAppealRepo) Create(tx *sqlx.Tx, pendingAppeal repository.PendingAppealDB) error {
	return nil
}
func (m *mockPendingAppealRepo) Update(tx *sqlx.Tx, pendingAppeal repository.PendingAppealDB) error {
	return nil
}
func (m *mockPendingAppealRepo) Delete(tx *sqlx.Tx, appealID int) error { return nil }

func (m *mockPendingAppealRepo) UpsertPendingAppealByID(appealID int, priority int64) error {
	if m.upsertFn != nil {
		return m.upsertFn(appealID, priority)
	}
	return nil
}
func (m *mockPendingAppealRepo) RemovePendingAppeal(appealID int) error {
	if m.removeFn != nil {
		return m.removeFn(appealID)
	}
	return nil
}

func TestSlotServiceFetchFreeSlotsEmpty(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "empty employee ids"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := testutil.NewMockDB(t)
			defer db.Close()
			svc := NewSlotService(db, nil)
			slots, err := svc.FetchFreeSlotsByEmployees(nil)
			require.NoError(t, err)
			assert.Empty(t, slots)
		})
	}
}

func TestEmployeeServiceGetAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get all employees"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := testutil.NewMockDB(t)
			defer db.Close()
			svc := NewEmployeeService(db, &mockEmployeeRepo{
				getAllFn: func() ([]model.Employee, error) { return []model.Employee{{ID: 1}}, nil },
			}, nil)
			items, err := svc.GetAll()
			require.NoError(t, err)
			assert.Len(t, items, 1)
		})
	}
}

type mockEmployeeRepo struct {
	getAllFn func() ([]model.Employee, error)
	getByIDFn func(id int) (model.Employee, error)
	createFn func(tx *sqlx.Tx, e model.Employee) (model.Employee, error)
}

func (m *mockEmployeeRepo) GetAll() ([]model.Employee, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockEmployeeRepo) GetByID(id int) (model.Employee, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return model.Employee{}, repository.ErrNotFound
}
func (m *mockEmployeeRepo) Create(tx *sqlx.Tx, e model.Employee) (model.Employee, error) {
	if m.createFn != nil {
		return m.createFn(tx, e)
	}
	e.ID = 1
	return e, nil
}
func (m *mockEmployeeRepo) Update(tx *sqlx.Tx, id int, e model.Employee) (model.Employee, error) {
	return e, nil
}
func (m *mockEmployeeRepo) Delete(tx *sqlx.Tx, id int) error { return nil }
func (m *mockEmployeeRepo) FetchAvailableEmployees(limit int) ([]repository.EmployeeWithAppealsCount, error) {
	return nil, nil
}
func (m *mockEmployeeRepo) GetEmployeeActiveAppeals(employeeID int) (int, error) { return 0, nil }

type mockSlotRepo struct {
	createFn      func(tx *sqlx.Tx, s model.Slot) (model.Slot, error)
	getCountFn    func(employeeID int) (int, error)
	getByAppealFn func(appealID int) (model.Slot, error)
}

func (m *mockSlotRepo) GetAll() ([]model.Slot, error) { return nil, nil }
func (m *mockSlotRepo) GetByID(id int) (model.Slot, error) { return model.Slot{}, nil }
func (m *mockSlotRepo) Create(tx *sqlx.Tx, s model.Slot) (model.Slot, error) {
	if m.createFn != nil {
		return m.createFn(tx, s)
	}
	return s, nil
}
func (m *mockSlotRepo) Update(tx *sqlx.Tx, id int, s model.Slot) (model.Slot, error) { return s, nil }
func (m *mockSlotRepo) Delete(tx *sqlx.Tx, id int) error { return nil }
func (m *mockSlotRepo) GetSlotsCount(employeeID int) (int, error) {
	if m.getCountFn != nil {
		return m.getCountFn(employeeID)
	}
	return 0, nil
}
func (m *mockSlotRepo) GetNeedToRemoveSlots(employeeID int) ([]model.Slot, error) { return nil, nil }
func (m *mockSlotRepo) SetNeedToRemoveValue(tx *sqlx.Tx, slot model.Slot, value bool) error { return nil }
func (m *mockSlotRepo) GetFreeSlots(employeeID int) ([]model.Slot, error) { return nil, nil }
func (m *mockSlotRepo) GetSlotByAppealID(appealID int) (model.Slot, error) {
	if m.getByAppealFn != nil {
		return m.getByAppealFn(appealID)
	}
	return model.Slot{}, sql.ErrNoRows
}
func (m *mockSlotRepo) GetNeedToRemoveSlot(employeeID int) (model.Slot, error) { return model.Slot{}, sql.ErrNoRows }
func (m *mockSlotRepo) GetRealSlots(employeeID int) ([]model.Slot, error) { return nil, nil }

func TestEmployeeServiceCreateWithSlots(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "create employee with slots"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			slotCreates := 0
			svc := NewEmployeeService(db, &mockEmployeeRepo{
				createFn: func(tx *sqlx.Tx, e model.Employee) (model.Employee, error) {
					e.ID = 7
					e.Limit = 2
					return e, nil
				},
			}, &mockSlotRepo{
				createFn: func(tx *sqlx.Tx, s model.Slot) (model.Slot, error) {
					slotCreates++
					return s, nil
				},
			})

			mock.ExpectBegin()
			mock.ExpectCommit()
			emp, err := svc.Create(model.Employee{Name: "E", Limit: 2})
			require.NoError(t, err)
			assert.Equal(t, 7, emp.ID)
			assert.Equal(t, 2, slotCreates)
		})
	}
}
