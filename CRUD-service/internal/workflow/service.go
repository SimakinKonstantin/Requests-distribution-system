package workflow

import "crud-service/internal/crud/repository"

type WorkflowService struct {
	repo repository.WorkflowRepository
}

func NewWorkflowService(repo repository.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (ws *WorkflowService) GetAll() ([]GetAllResponse, error) {
	dbWorkflows, err := ws.repo.GetAll()
	if err != nil {
		return nil, err
	}
	response := make([]GetAllResponse, len(dbWorkflows))
	for i, dbWorkflow := range dbWorkflows {
		response[i] = toGetAllResponse(dbWorkflow)
	}
	return response, nil
}

func (ws *WorkflowService) GetById(id int) (GetAllResponse, error) {
	dbWorkflow, err := ws.repo.GetByID(id)
	if err != nil {
		return GetAllResponse{}, err
	}
	return toGetAllResponse(dbWorkflow), nil
}

func toGetWorkflow(db repository.WorkflowDB) GetAllResponse {
	return Workflow{
		ID:     db.ID,
		Name:   db.Name,
		Status: db.Status,
		Data:   db.Data,
	}
}

func toGetAllResponse(db repository.GetAllDB) GetAllResponse {
	return GetAllResponse{
		ID:   db.ID,
		Name: db.Name,
	}
}
