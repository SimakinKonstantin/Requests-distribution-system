package workflow

import (
	"encoding/json"
	"testing"

	"crud-service/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowServiceGetAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get all workflows"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			mock.ExpectQuery(`SELECT id, name, status, data FROM workflows`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "data"}).
					AddRow(1, "wf1", "active", json.RawMessage(`{}`)))

			repo := NewWorkflowRepository(db)
			svc := NewWorkflowService(repo, nil)
			items, err := svc.GetAll()
			require.NoError(t, err)
			assert.Len(t, items, 1)
			assert.Equal(t, 1, items[0].ID)
			assert.Equal(t, "wf1", items[0].Name)
		})
	}
}

func TestWorkflowServiceAddAndGet(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "add and get workflow"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			wf := Workflow{Name: "test", Status: StatusLive, Nodes: []Node{}, Edges: []Edge{}}
			data, err := createWorkflowData(wf)
			require.NoError(t, err)

			mock.ExpectQuery(`INSERT INTO workflows`).
				WithArgs("test", "active", data).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

			repo := NewWorkflowRepository(db)
			svc := NewWorkflowService(repo, nil)
			created, err := svc.AddWorkflow(wf)
			require.NoError(t, err)
			assert.Equal(t, 42, created.ID)

			mock.ExpectQuery(`SELECT id, name, status, data FROM workflows WHERE id = \$1`).
				WithArgs(42).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "data"}).
					AddRow(42, "test", "active", data))

			got, err := svc.GetWorkflowById(42)
			require.NoError(t, err)
			assert.Equal(t, 42, got.ID)
			assert.Equal(t, "test", got.Name)
		})
	}
}

func TestWorkflowServicePauseResumeDeleteUpdate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "pause resume delete update"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			repo := NewWorkflowRepository(db)
			svc := NewWorkflowService(repo, nil)

			mock.ExpectExec(`UPDATE workflows SET status = \$1 WHERE id = \$2`).
				WithArgs(StatusPaused, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			require.NoError(t, svc.PauseWorkflow(1))

			mock.ExpectExec(`UPDATE workflows SET status = \$1 WHERE id = \$2`).
				WithArgs(StatusLive, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			require.NoError(t, svc.ResumeWorkflow(1))

			mock.ExpectExec(`DELETE FROM workflows WHERE id = \$1`).
				WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			require.NoError(t, svc.DeleteWorkflow(1))

			wf := Workflow{Name: "updated", Status: StatusLive}
			data, _ := createWorkflowData(wf)
			mock.ExpectExec(`UPDATE workflows SET name = \$1, status = \$2, data = \$3 WHERE id = \$4`).
				WithArgs("updated", "active", data, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			updated, err := svc.UpdateWorkflow(1, wf)
			require.NoError(t, err)
			assert.Equal(t, "updated", updated.Name)
		})
	}
}

func TestWorkflowServiceRun(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "run live workflows"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := testutil.NewMockDB(t)
			defer db.Close()

			attr := ThemeId
			cmp := Eq
			wfData, _ := json.Marshal(Workflow{
				Name:   "live",
				Status: StatusLive,
				Nodes: []Node{
					{ID: "p", Type: PredicateNode, Data: ptrAny(Predicate{Attribute: &attr, Comparison: &cmp, Values: []string{"1"}})},
				},
				Edges: []Edge{},
			})

			mock.ExpectQuery(`SELECT id, name, status, data FROM workflows`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "data"}).
					AddRow(1, "live", "active", wfData).
					AddRow(2, "paused", "paused", wfData))

			repo := NewWorkflowRepository(db)
			svc := NewWorkflowService(repo, nil)
			results := svc.Run(map[string]interface{}{"themeId": 1})
			assert.NotEmpty(t, results)
		})
	}
}

func TestCreateWorkflowData(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "create workflow data and db model"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := createWorkflowData(Workflow{Name: "x", Status: StatusLive})
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			db := createDbWorkflow("n", StatusLive, data)
			assert.Equal(t, "n", db.Name)
			assert.Equal(t, "active", db.Status)
		})
	}
}
