package balancer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAppealService struct {
	fetchFn      func(limit int) ([]service.PendingAppeal, error)
	assignFn     func(appealID, employeeID, slotID int) error
	upsertFn     func(appealID int) error
	isImportant  func(appealID int) bool
}

func (m *mockAppealService) GetAll() ([]model.Appeal, error)                       { return nil, nil }
func (m *mockAppealService) GetByID(id int) (model.Appeal, error)                  { return model.Appeal{}, nil }
func (m *mockAppealService) Create(a model.Appeal) (model.Appeal, error)           { return model.Appeal{}, nil }
func (m *mockAppealService) Update(id int, a model.Appeal) (model.Appeal, error) { return model.Appeal{}, nil }
func (m *mockAppealService) Delete(id int) error                                   { return nil }
func (m *mockAppealService) Close(id int) (model.Appeal, error)                    { return model.Appeal{}, nil }
func (m *mockAppealService) Assign(appealID, employeeID, slotID int) error {
	if m.assignFn != nil {
		return m.assignFn(appealID, employeeID, slotID)
	}
	return nil
}
func (m *mockAppealService) UpsertPendingAppealByID(appealID int) error {
	if m.upsertFn != nil {
		return m.upsertFn(appealID)
	}
	return nil
}
func (m *mockAppealService) FetchPendingAppeals(limit int) ([]service.PendingAppeal, error) {
	if m.fetchFn != nil {
		return m.fetchFn(limit)
	}
	return nil, nil
}
func (m *mockAppealService) IsImportant(appealID int) bool {
	if m.isImportant != nil {
		return m.isImportant(appealID)
	}
	return false
}

type mockEmployeeRepo struct {
	fetchFn    func(limit int) ([]repository.EmployeeWithAppealsCount, error)
	activeFn   func(employeeID int) (int, error)
}

func (m *mockEmployeeRepo) GetAll() ([]model.Employee, error) { return nil, nil }
func (m *mockEmployeeRepo) GetByID(id int) (model.Employee, error) {
	return model.Employee{}, nil
}
func (m *mockEmployeeRepo) Create(tx *sqlx.Tx, e model.Employee) (model.Employee, error) {
	return model.Employee{}, nil
}
func (m *mockEmployeeRepo) Update(tx *sqlx.Tx, id int, e model.Employee) (model.Employee, error) {
	return model.Employee{}, nil
}
func (m *mockEmployeeRepo) Delete(tx *sqlx.Tx, id int) error { return nil }
func (m *mockEmployeeRepo) FetchAvailableEmployees(limit int) ([]repository.EmployeeWithAppealsCount, error) {
	if m.fetchFn != nil {
		return m.fetchFn(limit)
	}
	return nil, nil
}
func (m *mockEmployeeRepo) GetEmployeeActiveAppeals(employeeID int) (int, error) {
	if m.activeFn != nil {
		return m.activeFn(employeeID)
	}
	return 0, nil
}

type mockSlotService struct {
	freeSlotsFn func(employeeIDs []int) (map[int][]model.Slot, error)
}

func (m *mockSlotService) GetAll() ([]model.Slot, error)             { return nil, nil }
func (m *mockSlotService) GetByID(id int) (model.Slot, error)        { return model.Slot{}, nil }
func (m *mockSlotService) Create(s model.Slot) (model.Slot, error)     { return model.Slot{}, nil }
func (m *mockSlotService) UpdateCount(id int, count int) error       { return nil }
func (m *mockSlotService) Delete(id int) error                       { return nil }
func (m *mockSlotService) FetchFreeSlotsByEmployees(employeeIDs []int) (map[int][]model.Slot, error) {
	if m.freeSlotsFn != nil {
		return m.freeSlotsFn(employeeIDs)
	}
	return map[int][]model.Slot{}, nil
}

func TestFindOptimalAssignments(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "single assignment"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamID := 1
			now := time.Now()
			appeals := []model.Appeal{{ID: 100, TeamID: &teamID}}
			employees := []model.Employee{{ID: 10, TeamIDs: []int{1}, Limit: 2}}
			freeSlots := map[int][]model.Slot{10: {{ID: 50, EmployeeID: 10, UpdatedAt: &now}}}

			m := &Matcher{
				appealService: &mockAppealService{},
				employeeRepo: &mockEmployeeRepo{
					activeFn: func(employeeID int) (int, error) { return 0, nil },
				},
			}

			assignments := m.FindOptimalAssignments(appeals, employees, freeSlots)
			require.Len(t, assignments, 1)
			assert.Equal(t, 100, assignments[0].AppealID)
			assert.Equal(t, 10, assignments[0].ManagerID)
			assert.Equal(t, 50, assignments[0].SlotID)
		})
	}
}

func TestFindOptimalAssignmentsNoTeam(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "appeal without team"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &Matcher{employeeRepo: &mockEmployeeRepo{activeFn: func(int) (int, error) { return 0, nil }}}
			assignments := m.FindOptimalAssignments([]model.Appeal{{ID: 1}}, nil, nil)
			assert.Empty(t, assignments)
		})
	}
}

func TestPickBestManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "pick least loaded manager"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			earlier := now.Add(-time.Hour)
			candidates := []*employeeState{
				{emploee: model.Employee{ID: 1}, activeAppealsCount: 1, lastAssign: &now},
				{emploee: model.Employee{ID: 2}, activeAppealsCount: 0, lastAssign: &earlier},
			}
			freeSlots := map[int][]model.Slot{
				1: {{ID: 10, EmployeeID: 1}},
				2: {{ID: 20, EmployeeID: 2}},
			}
			best := pickBestManager(candidates, freeSlots)
			require.NotNil(t, best)
			assert.Equal(t, 2, best.emploee.ID)
		})
	}
}

func TestHasFreeSlotAndPickOldest(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "has free slot and picks oldest"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &employeeState{emploee: model.Employee{ID: 1}, usedSlots: map[int]struct{}{10: {}}}
			freeSlots := map[int][]model.Slot{1: {{ID: 10}, {ID: 11}}}
			assert.True(t, hasFreeSlot(st, freeSlots))
			assert.Equal(t, 11, pickOldestFreeSlot(st, freeSlots))
			assert.Equal(t, 1, countFreeSlots(st, freeSlots))
		})
	}
}

func TestClassifyAppealPriority(t *testing.T) {
	tests := []struct {
		name      string
		important bool
		want      int
	}{
		{name: "important appeal", important: true, want: 0},
		{name: "regular appeal", important: false, want: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &Matcher{appealService: &mockAppealService{isImportant: func(int) bool { return tc.important }}}
			assert.Equal(t, tc.want, m.classifyAppealPriority(model.Appeal{ID: 1}))
		})
	}
}

func TestAssignerHandleAssignTask(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "assign task handled"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			assigner := NewAssigner(&mockAppealService{
				assignFn: func(appealID, employeeID, slotID int) error {
					called = true
					assert.Equal(t, 1, appealID)
					assert.Equal(t, 2, employeeID)
					assert.Equal(t, 3, slotID)
					return nil
				},
			})
			payload, err := json.Marshal(AssignPayload{AppealID: 1, ManagerID: 2, SlotID: 3})
			require.NoError(t, err)
			task := asynq.NewTask(TaskTypeAssign, payload)
			require.NoError(t, assigner.HandleAssignTask(context.Background(), task))
			assert.True(t, called)
		})
	}
}

func TestAssignerAlreadyAssigned(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "already assigned ignored"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assigner := NewAssigner(&mockAppealService{
				assignFn: func(int, int, int) error { return service.ErrAppealAlreadyAssigned },
			})
			payload, _ := json.Marshal(AssignPayload{AppealID: 1, ManagerID: 2, SlotID: 3})
			require.NoError(t, assigner.HandleAssignTask(context.Background(), asynq.NewTask(TaskTypeAssign, payload)))
		})
	}
}

func TestAssignerSlotOccupied(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "slot occupied ignored"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assigner := NewAssigner(&mockAppealService{
				assignFn: func(int, int, int) error { return sql.ErrNoRows },
			})
			payload, _ := json.Marshal(AssignPayload{AppealID: 1, ManagerID: 2, SlotID: 3})
			require.NoError(t, assigner.HandleAssignTask(context.Background(), asynq.NewTask(TaskTypeAssign, payload)))
		})
	}
}

func TestAssignerInvalidPayload(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "invalid payload"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assigner := NewAssigner(&mockAppealService{})
			err := assigner.HandleAssignTask(context.Background(), asynq.NewTask(TaskTypeAssign, []byte("bad")))
			assert.Error(t, err)
		})
	}
}

func TestBalancerUpdateService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "batch update calls upsert"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := 0
			svc := NewBalancerUpdateService(&mockAppealService{
				upsertFn: func(appealID int) error {
					called++
					assert.Equal(t, 5, appealID)
					return nil
				},
			}, nil)

			payload, err := json.Marshal(BatchUpdatePayload{
				Type: BatchDistributionRequests,
				Events: []ProcessedEvent{{
					RabbitEvent: RabbitEvent{
						Name: EventAppealNeedsDistribution,
						Payload: RabbitEventBody{AppealID: 5, TeamID: 1},
					},
				}},
			})
			require.NoError(t, err)
			require.NoError(t, svc.HandleBatchUpdateTask(context.Background(), asynq.NewTask(TaskTypeBatchUpdate, payload)))
			assert.Equal(t, 1, called)
		})
	}
}

func TestBalancerUpdateServiceUpsertError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "upsert returns error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewBalancerUpdateService(&mockAppealService{
				upsertFn: func(int) error { return errors.New("upsert failed") },
			}, nil)
			payload, _ := json.Marshal(BatchUpdatePayload{
				Type: BatchDistributionRequests,
				Events: []ProcessedEvent{{
					RabbitEvent: RabbitEvent{
						Name: EventAppealNeedsDistribution,
						Payload: RabbitEventBody{AppealID: 1},
					},
				}},
			})
			err := svc.HandleBatchUpdateTask(context.Background(), asynq.NewTask(TaskTypeBatchUpdate, payload))
			assert.Error(t, err)
		})
	}
}

func TestBalancerUpdateServiceUnknownBatchType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "unknown type ignored"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewBalancerUpdateService(&mockAppealService{}, nil)
			payload, _ := json.Marshal(BatchUpdatePayload{Type: "unknown"})
			require.NoError(t, svc.HandleBatchUpdateTask(context.Background(), asynq.NewTask(TaskTypeBatchUpdate, payload)))
		})
	}
}

func TestAsynqTasks(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "asynq task constructors"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := NewDistributionTickTask()
			assert.Equal(t, TaskTypeDistributionTick, task.Type())

			assignTask, err := NewAssignTask(AssignPayload{AppealID: 1})
			require.NoError(t, err)
			assert.Equal(t, TaskTypeAssign, assignTask.Type())

			batchTask, err := NewBatchUpdateTask(BatchUpdatePayload{Type: BatchDistributionRequests})
			require.NoError(t, err)
			assert.Equal(t, TaskTypeBatchUpdate, batchTask.Type())
		})
	}
}

func TestGroupEventsByType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "group distribution events"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := []ProcessedEvent{{
				RabbitEvent: RabbitEvent{Name: EventAppealNeedsDistribution, Payload: RabbitEventBody{AppealID: 1}},
			}}
			grouped := groupEventsByType(events)
			assert.Len(t, grouped[BatchDistributionRequests], 1)
		})
	}
}

func TestQueueForBatchType(t *testing.T) {
	tests := []struct {
		name string
		in   JobBatchType
		want string
	}{
		{name: "high", in: BatchAppealStatusChanges, want: "state-high"},
		{name: "medium", in: BatchManagerAvailabilityChange, want: "state-medium"},
		{name: "low", in: BatchDistributionRequests, want: "state-low"},
		{name: "default", in: "other", want: "state-low"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, queueForBatchType(tc.in))
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name string
		msg  string
	}{
		{name: "error text", msg: "bad role"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := &ConfigError{Msg: tc.msg}
			assert.Equal(t, tc.msg, err.Error())
		})
	}
}

func TestEventHandlerCreateJobName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "job name from event"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewEventHandlerService(Config{BatchSize: 10}, nil)
			name := svc.createJobName(ProcessedEvent{
				RabbitEvent: RabbitEvent{Name: EventAppealNeedsDistribution, Payload: RabbitEventBody{AppealID: 7}},
			})
			assert.Equal(t, "APPEAL_NEEDS_DISTRIBUTION_7", name)
		})
	}
}

func TestEventHandlerAddEventToBatchNoFlush(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "batch accumulates"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewEventHandlerService(Config{BatchSize: 10, BatchTimeout: time.Second}, nil)
			svc.addEventToBatch(context.Background(), ProcessedEvent{
				RabbitEvent: RabbitEvent{Name: EventAppealNeedsDistribution, Payload: RabbitEventBody{AppealID: 1}},
			})
			assert.Len(t, svc.batch, 1)
		})
	}
}

func TestEventHandlerHandleDelivery(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "delivery handled"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewEventHandlerService(Config{BatchSize: 100}, nil)
			body, _ := json.Marshal(RabbitEvent{Name: EventAppealNeedsDistribution, Payload: RabbitEventBody{AppealID: 3}})
			require.NoError(t, svc.handleDelivery(context.Background(), amqp.Delivery{Body: body}))
			assert.Len(t, svc.batch, 1)
		})
	}
}

func TestMatcherHandleDistributionTickEmpty(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "empty sources produce no error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &Matcher{
				appealService: &mockAppealService{fetchFn: func(int) ([]service.PendingAppeal, error) { return nil, nil }},
				employeeRepo:  &mockEmployeeRepo{fetchFn: func(int) ([]repository.EmployeeWithAppealsCount, error) { return nil, nil }},
				slotService:   &mockSlotService{},
				cfg:           Config{FetchAppealsLimit: 10, FetchManagersLimit: 10},
			}
			require.NoError(t, m.HandleDistributionTick(context.Background(), NewDistributionTickTask()))
		})
	}
}
