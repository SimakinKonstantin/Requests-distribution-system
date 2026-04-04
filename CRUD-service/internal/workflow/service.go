package workflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

type WorkflowService struct {
	repo         WorkflowRepository
	teamAssigner TeamAssigner
}

func NewWorkflowService(repo *WorkflowRepository, teamAssigner TeamAssigner) WorkflowService {
	return WorkflowService{repo: *repo, teamAssigner: teamAssigner}
}

func (ws *WorkflowService) GetAll() ([]GetAllResponse, error) {
	dbWorkflows, err := ws.repo.All()
	if err != nil {
		return nil, err
	}
	response := make([]GetAllResponse, len(dbWorkflows))

	for i := 0; i < len(dbWorkflows); i++ {
		response[i].ID = dbWorkflows[i].ID
		response[i].Name = dbWorkflows[i].Name
	}
	return response, nil
}

// func (ws *WorkflowService) GetById(id int) (GetAllResponse, error) {
// 	dbWorkflow, err := ws.repo.GetByID(id)
// 	if err != nil {
// 		return GetAllResponse{}, err
// 	}
// 	return toGetAllResponse(dbWorkflow), nil
// }

func (ws *WorkflowService) Run(payload map[string]interface{}) []BlockResult {
	var (
		wg      sync.WaitGroup
		results []BlockResult
	)

	slog.Info("Starting workflow")

	dbWorkflows, err := ws.repo.All()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get all workflows from db: %w", err))
		return nil
	}

	slog.Info("Got all workflows from db")

	var executors []*Executor
	for _, w := range dbWorkflows {
		var workflow Workflow
		if err := json.Unmarshal(w.Data, &workflow); err != nil {
			slog.Error(fmt.Sprintf("failed to unmarshal workflow %s: %w", w.ID, err))
			continue
		}
		workflow.ID = w.ID

		if w.Status != string(StatusLive) {
			continue
		}

		executor := ws.buildExecutor(workflow)
		executors = append(executors, executor)
	}

	slog.Info("Built executors")

	for _, executor := range executors {
		wg.Add(1)
		go func(e *Executor) {
			defer wg.Done()
			res := e.ExecuteWorkflow(payload)
			results = append(results, res)
		}(executor)
	}

	wg.Wait()

	slog.Info("Finished executing workflows")

	return results
}

func (wm *WorkflowService) AddWorkflow(workflow Workflow) (Workflow, error) {
	data, err := createWorkflowData(workflow)
	if err != nil {
		return Workflow{}, fmt.Errorf("failed to create workflow data: %w", err)
	}

	dbWorkflow := createDbWorkflow(
		workflow.Name,
		workflow.Status,
		data,
	)

	slog.Info(fmt.Sprintf("dbWorkflow: %+v", dbWorkflow))

	ID, err := wm.repo.Save(dbWorkflow)
	if err != nil {
		return Workflow{}, fmt.Errorf("failed to save workflow: %w", err)
	}

	workflow.ID = ID
	return workflow, nil
}

func (wm *WorkflowService) PauseWorkflow(id int) error {
	if err := wm.repo.UpdateStatus(id, StatusPaused); err != nil {
		return fmt.Errorf("failed to pause workflow: %w", err)
	}

	return nil
}

func (wm *WorkflowService) ResumeWorkflow(id int) error {
	if err := wm.repo.UpdateStatus(id, StatusLive); err != nil {
		return fmt.Errorf("failed to resume workflow: %w", err)
	}

	return nil
}

func (wm *WorkflowService) GetWorkflowById(ID int) (Workflow, error) {
	dwWorkflow, err := wm.repo.Get(ID)
	if err != nil {
		return Workflow{}, err
	}

	var workflow Workflow
	if err := json.Unmarshal(dwWorkflow.Data, &workflow); err != nil {
		return Workflow{}, fmt.Errorf("failed to Unmarshal workflow: %w", err)
	}
	workflow.ID = dwWorkflow.ID
	workflow.Name = dwWorkflow.Name
	workflow.Status = Status(dwWorkflow.Status)

	return workflow, nil
}

func (wm *WorkflowService) DeleteWorkflow(ID int) error {
	if err := wm.repo.Delete(ID); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	return nil
}

func (wm *WorkflowService) UpdateWorkflow(ID int, newWorkflow Workflow) (Workflow, error) {
	data, err := createWorkflowData(newWorkflow)
	if err != nil {
		return Workflow{}, fmt.Errorf("failed to create workflow data: %w", err)
	}
	dbWorkflow := WorkflowDB{
		ID:     ID,
		Name:   newWorkflow.Name,
		Status: string(newWorkflow.Status),
		Data:   data,
	}
	if err = wm.repo.Update(ID, dbWorkflow); err != nil {
		return Workflow{}, fmt.Errorf("failed to update workflow: %w", err)
	}
	return newWorkflow, nil
}

// func (wm *WorkflowService) GetWorkflows() ([]Workflow, error) {
// 	dbWorkflows, err := wm.repo.All()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var results []Workflow
// 	for _, dbWorkflow := range dbWorkflows {
// 		dto := Workflow{
// 			ID:     dbWorkflow.ID,
// 			Name:   dbWorkflow.Name,
// 			Status: Status(dbWorkflow.Status),
// 		}

// 		dto := WorkflowGetManyDto{
// 			CreatedAt: dbWorkflow.CreatedAt.Format(time.RFC3339),
// 			ID:        dbWorkflow.ID,
// 			Name:      dbWorkflow.Name,
// 			Status:    WorkflowDtoStatus(dbWorkflow.Status),
// 		}

// 		results = append(results, dto)
// 	}

// 	//TODO: DELETE PAGINATION
// 	return gen.WorkflowGetManyResponse{
// 		Results: results,
// 		Meta: gen.WorkflowGetManyMeta{
// 			HasMore:    false,
// 			NextCursor: "",
// 			Total:      len(results),
// 		},
// 	}, nil
// }

func (wm *WorkflowService) buildExecutor(workflow Workflow) *Executor {
	nextMap := buildNextMap(workflow.Edges)
	startId := findStartNode(workflow.Nodes, nextMap)
	chain := wm.buildChain(workflow.Nodes, nextMap, startId)
	executor := newWorkflowExecutor(chain, Status(workflow.Status))
	return executor
}

func createDbWorkflow(name string, status Status, data json.RawMessage) WorkflowDB {
	return WorkflowDB{
		Name:   name,
		Status: string(status),
		Data:   data,
	}
}

func createWorkflowData(workflow Workflow) (json.RawMessage, error) {
	bytesData, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}
	return bytesData, nil
}
