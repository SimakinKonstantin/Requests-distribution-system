package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"
	"crud-service/internal/testutil"
	"crud-service/internal/workflow"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"
)

type stubEmployees struct {
	service.EmployeeService
	getAll func() ([]model.Employee, error)
	getBy  func(id int) (model.Employee, error)
	create func(e model.Employee) (model.Employee, error)
	update func(id int, e model.Employee) (model.Employee, error)
	del    func(id int) error
}

func (s stubEmployees) GetAll() ([]model.Employee, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubEmployees) GetByID(id int) (model.Employee, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Employee{}, nil
}
func (s stubEmployees) Create(e model.Employee) (model.Employee, error) {
	if s.create != nil {
		return s.create(e)
	}
	return e, nil
}
func (s stubEmployees) Update(id int, e model.Employee) (model.Employee, error) {
	if s.update != nil {
		return s.update(id, e)
	}
	return e, nil
}
func (s stubEmployees) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}

type stubSlots struct {
	service.SlotService
	getAll       func() ([]model.Slot, error)
	getBy        func(id int) (model.Slot, error)
	create       func(s model.Slot) (model.Slot, error)
	del          func(id int) error
	updateCount  func(id int, count int) error
}

func (s stubSlots) GetAll() ([]model.Slot, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubSlots) GetByID(id int) (model.Slot, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Slot{}, nil
}
func (s stubSlots) Create(slot model.Slot) (model.Slot, error) {
	if s.create != nil {
		return s.create(slot)
	}
	return slot, nil
}
func (s stubSlots) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}
func (s stubSlots) UpdateCount(id int, count int) error {
	if s.updateCount != nil {
		return s.updateCount(id, count)
	}
	return nil
}
func (s stubSlots) FetchFreeSlotsByEmployees([]int) (map[int][]model.Slot, error) {
	return nil, nil
}

type stubAppeals struct {
	service.AppealService
	getAll func() ([]model.Appeal, error)
	getBy  func(id int) (model.Appeal, error)
	create func(a model.Appeal) (model.Appeal, error)
	update func(id int, a model.Appeal) (model.Appeal, error)
	del    func(id int) error
	close  func(id int) (model.Appeal, error)
}

func (s stubAppeals) GetAll() ([]model.Appeal, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubAppeals) GetByID(id int) (model.Appeal, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Appeal{}, nil
}
func (s stubAppeals) Create(a model.Appeal) (model.Appeal, error) {
	if s.create != nil {
		return s.create(a)
	}
	return a, nil
}
func (s stubAppeals) Update(id int, a model.Appeal) (model.Appeal, error) {
	if s.update != nil {
		return s.update(id, a)
	}
	return a, nil
}
func (s stubAppeals) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}
func (s stubAppeals) Close(id int) (model.Appeal, error) {
	if s.close != nil {
		return s.close(id)
	}
	return model.Appeal{ID: id, Status: "closed"}, nil
}
func (s stubAppeals) Assign(int, int, int) error                         { return nil }
func (s stubAppeals) UpsertPendingAppealByID(int) error                  { return nil }
func (s stubAppeals) FetchPendingAppeals(int) ([]service.PendingAppeal, error) { return nil, nil }
func (s stubAppeals) IsImportant(int) bool                               { return false }

type stubClients struct {
	service.ClientService
	getAll   func() ([]model.Client, error)
	getBy    func(id int) (model.Client, error)
	create   func(c model.Client) (model.Client, error)
	update   func(id int, c model.Client) (model.Client, error)
	del      func(id int) error
	getEmails func() ([]string, error)
}

func (s stubClients) GetAll() ([]model.Client, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubClients) GetByID(id int) (model.Client, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Client{}, nil
}
func (s stubClients) Create(c model.Client) (model.Client, error) {
	if s.create != nil {
		return s.create(c)
	}
	return c, nil
}
func (s stubClients) Update(id int, c model.Client) (model.Client, error) {
	if s.update != nil {
		return s.update(id, c)
	}
	return c, nil
}
func (s stubClients) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}
func (s stubClients) GetEmails() ([]string, error) {
	if s.getEmails != nil {
		return s.getEmails()
	}
	return nil, nil
}

type stubCRUD[T any] struct {
	getAll func() ([]T, error)
	getBy  func(id int) (T, error)
	create func(v T) (T, error)
	update func(id int, v T) (T, error)
	del    func(id int) error
}

type stubSubthemes struct{ stubCRUD[model.Subtheme] }

func (s stubSubthemes) GetAll() ([]model.Subtheme, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubSubthemes) GetByID(id int) (model.Subtheme, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Subtheme{}, nil
}
func (s stubSubthemes) Create(v model.Subtheme) (model.Subtheme, error) {
	if s.create != nil {
		return s.create(v)
	}
	return v, nil
}
func (s stubSubthemes) Update(id int, v model.Subtheme) (model.Subtheme, error) {
	if s.update != nil {
		return s.update(id, v)
	}
	return v, nil
}
func (s stubSubthemes) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}

type stubThemes struct{ stubCRUD[model.Theme] }

func (s stubThemes) GetAll() ([]model.Theme, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubThemes) GetByID(id int) (model.Theme, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Theme{}, nil
}
func (s stubThemes) Create(v model.Theme) (model.Theme, error) {
	if s.create != nil {
		return s.create(v)
	}
	return v, nil
}
func (s stubThemes) Update(id int, v model.Theme) (model.Theme, error) {
	if s.update != nil {
		return s.update(id, v)
	}
	return v, nil
}
func (s stubThemes) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}

type stubTeams struct{ stubCRUD[model.Team] }

func (s stubTeams) GetAll() ([]model.Team, error) {
	if s.getAll != nil {
		return s.getAll()
	}
	return nil, nil
}
func (s stubTeams) GetByID(id int) (model.Team, error) {
	if s.getBy != nil {
		return s.getBy(id)
	}
	return model.Team{}, nil
}
func (s stubTeams) Create(v model.Team) (model.Team, error) {
	if s.create != nil {
		return s.create(v)
	}
	return v, nil
}
func (s stubTeams) Update(id int, v model.Team) (model.Team, error) {
	if s.update != nil {
		return s.update(id, v)
	}
	return v, nil
}
func (s stubTeams) Delete(id int) error {
	if s.del != nil {
		return s.del(id)
	}
	return nil
}
func (s stubTeams) AssignTeam(int, int) error { return nil }
func (s stubTeams) GetTeam(int, *int, bool) (model.Team, error) {
	return model.Team{}, nil
}

func emptyWorkflowService(t *testing.T) workflow.WorkflowService {
	t.Helper()
	db, _ := testutil.NewMockDB(t)
	t.Cleanup(func() { _ = db.Close() })
	repo := workflow.NewWorkflowRepository(db)
	return workflow.NewWorkflowService(repo, nil)
}

func newTestHandler(t *testing.T) *Handler {
	return New(
		stubEmployees{},
		stubSlots{},
		stubAppeals{},
		stubSubthemes{},
		stubClients{},
		stubThemes{},
		stubTeams{},
		emptyWorkflowService(t),
	)
}

func doRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestHealthAndCORS(t *testing.T) {
	h := newTestHandler(t).InitRoutes()
	w := doRequest(h, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	w = doRequest(h, http.MethodOptions, "/employees", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestEmployeesHandler(t *testing.T) {
	h := New(
		stubEmployees{
			getAll: func() ([]model.Employee, error) { return []model.Employee{{ID: 1, Name: "A"}}, nil },
			getBy:  func(id int) (model.Employee, error) { return model.Employee{ID: id}, nil },
			create: func(e model.Employee) (model.Employee, error) { e.ID = 2; return e, nil },
			update: func(id int, e model.Employee) (model.Employee, error) { e.ID = id; e.Limit = 3; return e, nil },
			del:    func(id int) error { return nil },
		},
		stubSlots{updateCount: func(id int, count int) error { return nil }},
		stubAppeals{}, stubSubthemes{}, stubClients{}, stubThemes{}, stubTeams{}, emptyWorkflowService(t),
	).InitRoutes()

	w := doRequest(h, http.MethodGet, "/employees", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodPost, "/employees", model.Employee{Name: "New"})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = doRequest(h, http.MethodGet, "/employees/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodPut, "/employees/1", model.Employee{Name: "Upd"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodDelete, "/employees/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)

	w = doRequest(h, http.MethodGet, "/employees/bad", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doRequest(h, http.MethodPatch, "/employees", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestEmployeesNotFound(t *testing.T) {
	h := New(
		stubEmployees{getBy: func(id int) (model.Employee, error) { return model.Employee{}, repository.ErrNotFound }},
		stubSlots{}, stubAppeals{}, stubSubthemes{}, stubClients{}, stubThemes{}, stubTeams{}, emptyWorkflowService(t),
	).InitRoutes()
	w := doRequest(h, http.MethodGet, "/employees/99", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAppealsClose(t *testing.T) {
	h := New(
		stubEmployees{}, stubSlots{},
		stubAppeals{close: func(id int) (model.Appeal, error) { return model.Appeal{ID: id, Status: "closed"}, nil }},
		stubSubthemes{}, stubClients{}, stubThemes{}, stubTeams{}, emptyWorkflowService(t),
	).InitRoutes()

	w := doRequest(h, http.MethodPost, "/appeals/5/close", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodPost, "/appeals/x/close", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientsEmails(t *testing.T) {
	h := New(
		stubEmployees{}, stubSlots{}, stubAppeals{}, stubSubthemes{},
		stubClients{getEmails: func() ([]string, error) { return []string{"a@b.c"}, nil }},
		stubThemes{}, stubTeams{}, emptyWorkflowService(t),
	).InitRoutes()

	w := doRequest(h, http.MethodGet, "/clients/emails", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "a@b.c")
}

func TestBadJSON(t *testing.T) {
	h := newTestHandler(t).InitRoutes()
	req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceError(t *testing.T) {
	h := New(
		stubEmployees{getAll: func() ([]model.Employee, error) { return nil, errors.New("db down") }},
		stubSlots{}, stubAppeals{}, stubSubthemes{}, stubClients{}, stubThemes{}, stubTeams{}, emptyWorkflowService(t),
	).InitRoutes()
	w := doRequest(h, http.MethodGet, "/employees", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWorkflowsHandler(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	defer db.Close()
	repo := workflow.NewWorkflowRepository(db)
	wfSvc := workflow.NewWorkflowService(repo, nil)

	mock.ExpectQuery(`SELECT id, name, status, data FROM workflows`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "data"}).
			AddRow(1, "wf", "active", []byte(`{"name":"wf","status":"active","nodes":[],"edges":[]}`)))

	data := []byte(`{"name":"new","status":"active","nodes":[],"edges":[]}`)
	mock.ExpectQuery(`INSERT INTO workflows`).
		WithArgs("new", "active", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	mock.ExpectQuery(`SELECT id, name, status, data FROM workflows WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "data"}).
			AddRow(1, "wf", "active", data))

	mock.ExpectExec(`UPDATE workflows SET name = \$1, status = \$2, data = \$3 WHERE id = \$4`).
		WithArgs("upd", "", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM workflows WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	h := New(
		stubEmployees{}, stubSlots{}, stubAppeals{}, stubSubthemes{}, stubClients{}, stubThemes{}, stubTeams{},
		wfSvc,
	).InitRoutes()

	w := doRequest(h, http.MethodGet, "/workflows", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodPost, "/workflows", workflow.Workflow{Name: "new", Status: workflow.StatusLive})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = doRequest(h, http.MethodGet, "/workflows/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodPut, "/workflows/1", workflow.Workflow{Name: "upd"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doRequest(h, http.MethodDelete, "/workflows/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSlotsThemesTeamsSubthemesCRUD(t *testing.T) {
	h := New(
		stubEmployees{}, stubSlots{
			getAll: func() ([]model.Slot, error) { return []model.Slot{{ID: 1}}, nil },
			getBy:  func(id int) (model.Slot, error) { return model.Slot{ID: id}, nil },
			create: func(s model.Slot) (model.Slot, error) { s.ID = 2; return s, nil },
		},
		stubAppeals{},
		stubSubthemes{stubCRUD: stubCRUD[model.Subtheme]{
			getAll: func() ([]model.Subtheme, error) { return []model.Subtheme{{ID: 1}}, nil },
			create: func(v model.Subtheme) (model.Subtheme, error) { v.ID = 2; return v, nil },
		}},
		stubClients{},
		stubThemes{stubCRUD: stubCRUD[model.Theme]{
			getAll: func() ([]model.Theme, error) { return []model.Theme{{ID: 1}}, nil },
		}},
		stubTeams{stubCRUD: stubCRUD[model.Team]{
			getAll: func() ([]model.Team, error) { return []model.Team{{ID: 1}}, nil },
		}},
		emptyWorkflowService(t),
	).InitRoutes()

	for _, tc := range []struct{ method, path string; code int }{
		{http.MethodGet, "/slots", http.StatusOK},
		{http.MethodPost, "/slots", http.StatusCreated},
		{http.MethodGet, "/slots/1", http.StatusOK},
		{http.MethodGet, "/subthemes", http.StatusOK},
		{http.MethodPost, "/subthemes", http.StatusCreated},
		{http.MethodGet, "/themes", http.StatusOK},
		{http.MethodGet, "/teams", http.StatusOK},
	} {
		t.Run(tc.path, func(t *testing.T) {
			body := map[string]any{"employeeId": 1}
			if tc.method == http.MethodPost && tc.path == "/subthemes" {
				body = map[string]any{"name": "s", "themeId": 1}
			}
			w := doRequest(h, tc.method, tc.path, body)
			assert.Equal(t, tc.code, w.Code)
		})
	}
}

func TestParseID(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/employees/abc", nil)
	_, ok := parseID(w, r, "/employees/")
	assert.False(t, ok)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotFoundOrInternal(t *testing.T) {
	w := httptest.NewRecorder()
	notFoundOrInternal(w, repository.ErrNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	notFoundOrInternal(w, errors.New("boom"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"k": "v"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"k":"v"`)
}
