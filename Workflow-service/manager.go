package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"workflow-service/gen"

	"github.com/jackc/pgx/v5/pgconn"
)

type workflowRepository interface {
	Save(workflow models.DbWorkflow) (string, error)
	All() ([]models.DbWorkflow, error)
	Update(id string, workflow models.DbWorkflow) error
	UpdateStatus(id string, state gen.WorkflowStatus) error
	Get(id string) (models.DbWorkflow, error)
	Author(profileId string) (models.DbAuthor, error)
}

type clientRepository interface {
	GetEmailsByCompanyId(ctx context.Context, companyId string) ([]string, error)
}

var (
	ErrorWorkflowExists = errors.New("workflow already exists")
)

type Manager struct {
	logger             slog.Logger
	rabbitConnection   string
	workflowRepository workflowRepository
	// backEndClient      *genclient.ClientWithResponses
	// apiSmartChatClient *apismartchat.Service
	// aiRepository       repository.AiRepository
	// clientRepository   clientRepository
}

func NewManager(
	logger slog.Logger,
	rabbitConnection string,
	dbWorkflow workflowRepository,
	// backEndClient *genclient.ClientWithResponses,
	// apiSmartChatClient *apismartchat.Service,
	// aiRepository repository.AiRepository,
	clientRepository clientRepository,
) *Manager {
	return &Manager{
		logger:             logger,
		rabbitConnection:   rabbitConnection,
		workflowRepository: dbWorkflow,
		// backEndClient:      backEndClient,
		// apiSmartChatClient: apiSmartChatClient,
		// aiRepository:       aiRepository,
		// clientRepository:   clientRepository,
	}
}

func (wm *Manager) Run(payload map[string]interface{}) []actionBlockResult {
	var (
		wg      sync.WaitGroup
		results []actionBlockResult
	)

	// teamId, ok := payload["teamId"].(string)
	// if !ok {
	// 	wm.logger.Errorf("failed to get teamId from payload")
	// 	return nil
	// }

	// _, err := wm.aiRepository.GetAiTeamInfo(teamId)
	// switch {
	// case err == nil:
	// 	wm.logger.Debug("workflows don't work for AI chat")
	// 	return nil
	// case !errors.Is(err, sql.ErrNoRows):
	// 	wm.logger.Errorf("error checking team: %s", err)
	// 	return nil
	// }

	// dbWorkflows, err := wm.workflowRepository.All()
	// if err != nil {
	// 	wm.logger.Errorf("failed to get all workflows from db: %w", err)
	// 	return nil
	// }

	// var executors []*Executor
	// for _, w := range dbWorkflows {
	// 	var workflow gen.Workflow
	// 	if err := json.Unmarshal(w.Data, &workflow); err != nil {
	// 		wm.logger.Errorf("failed to unmarshal workflow %s: %w", w.Id, err)
	// 		continue
	// 	}
	// 	workflow.Id = w.Id

	// 	if w.Status != string(gen.Live) {
	// 		continue
	// 	}

	// 	executor := wm.buildExecutor(workflow)
	// 	executors = append(executors, executor)
	// }

	// for _, executor := range executors {
	// 	wg.Add(1)
	// 	go func(e *Executor) {
	// 		defer wg.Done()
	// 		res := e.ExecuteWorkflow(payload)
	// 		results = append(results, res)
	// 	}(executor)
	// }

	// wg.Wait()
	// return results
	return nil
}

func (wm *Manager) AddWorkflow(workflowDto gen.WorkflowCreateOneDto) (gen.Workflow, error) {
	createdAt := time.Now()
	author, err := wm.workflowRepository.Author(workflowDto.AuthorProfileId)
	if err != nil {
		return gen.Workflow{}, err
	}

	respAuthor := createAuthorResponse(author)

	createdWorkflow, err := createWorkflowFromDto(workflowDto, respAuthor, createdAt)
	if err != nil {
		return gen.Workflow{}, err
	}

	data, err := createWorkflowData(createdWorkflow)
	if err != nil {
		return gen.Workflow{}, err
	}

	dbWorkflow := createDbWorkflow(
		workflowDto.Name,
		workflowDto.Status,
		data,
		workflowDto.AuthorProfileId,
		createdAt,
		createdAt,
	)

	id, err := wm.workflowRepository.Save(dbWorkflow)
	if err != nil {
		return gen.Workflow{}, wm.handleWorkflowError(id, err)
	}

	createdWorkflow.Id = id
	return createdWorkflow, nil
}

func (wm *Manager) PauseWorkflow(id string) error {
	if err := wm.workflowRepository.UpdateStatus(id, gen.Paused); err != nil {
		return fmt.Errorf("failed to pause workflow: %w", err)
	}

	return nil
}

func (wm *Manager) ResumeWorkflow(id string) error {
	if err := wm.workflowRepository.UpdateStatus(id, gen.Live); err != nil {
		return fmt.Errorf("failed to resume workflow: %w", err)
	}

	return nil
}

func (wm *Manager) RemoveWorkflow(id string) error {
	if err := wm.workflowRepository.UpdateStatus(id, gen.Archived); err != nil {
		return fmt.Errorf("failed to remove workflow: %w", err)
	}

	return nil
}

func (wm *Manager) GetWorkflowById(id string) (gen.Workflow, error) {
	dwWorkflow, err := wm.workflowRepository.Get(id)
	if err != nil {
		return gen.Workflow{}, err
	}

	var workflow gen.Workflow
	if err := json.Unmarshal(dwWorkflow.Data, &workflow); err != nil {
		return gen.Workflow{}, fmt.Errorf("failed to Unmarshal workflow: %w", err)
	}
	workflow.Id = dwWorkflow.Id

	return workflow, nil
}

func (wm *Manager) GetWorkflows() (gen.WorkflowGetManyResponse, error) {
	// dbWorkflows, err := wm.workflowRepository.All()
	// if err != nil {
	// 	return gen.WorkflowGetManyResponse{}, err
	// }

	// var results []gen.WorkflowGetManyDto
	// for _, dbWorkflow := range dbWorkflows {
	// 	dto := gen.WorkflowGetManyDto{
	// 		CreatedAt: dbWorkflow.CreatedAt.Format(time.RFC3339),
	// 		Id:        dbWorkflow.Id,
	// 		Name:      dbWorkflow.Name,
	// 		Status:    gen.WorkflowDtoStatus(dbWorkflow.Status),
	// 	}

	// 	results = append(results, dto)
	// }

	// return gen.WorkflowGetManyResponse{
	// 	Results: results,
	// 	Meta: gen.WorkflowGetManyMeta{
	// 		HasMore:    false,
	// 		NextCursor: "",
	// 		Total:      len(results),
	// 	},
	// }, nil
}

func (wm *Manager) UpdateWorkflow(id string, newWorkflow gen.WorkflowUpdateOneDto) (gen.Workflow, error) {
	// author, err := wm.workflowRepository.Author(newWorkflow.AuthorProfileId)
	// if err != nil {
	// 	return gen.Workflow{}, err
	// }

	// respAuthor := createAuthorResponse(author)

	// updatedAt := time.Now()
	// updatedWorkflow, err := createWorkflowFromDto(newWorkflow, respAuthor, updatedAt)
	// if err != nil {
	// 	return gen.Workflow{}, err
	// }

	// data, err := createWorkflowData(updatedWorkflow)
	// if err != nil {
	// 	return gen.Workflow{}, err
	// }

	// dbWorkflow := models.DbWorkflow{
	// 	Name:           newWorkflow.Name,
	// 	Status:         string(newWorkflow.Status),
	// 	Data:           data,
	// 	AuthorProfleId: newWorkflow.AuthorProfileId,
	// 	UpdatedAt:      updatedAt,
	// }

	// if newWorkflow.Status == gen.Archived {
	// 	archivedAt := time.Now()
	// 	dbWorkflow.ArchivedAt = &archivedAt
	// }

	// if err = wm.workflowRepository.Update(id, dbWorkflow); err != nil {
	// 	return gen.Workflow{}, wm.handleWorkflowError(id, err)
	// }

	// return updatedWorkflow, nil
}

func (wm *Manager) buildExecutor(workflow gen.Workflow) *Executor {
	// nextMap := buildNextMap(workflow.Edges)
	// startId := findStartNode(workflow.Nodes, nextMap)
	// chain := wm.buildChain(workflow.Nodes, nextMap, startId)
	// executor := newWorkflowExecutor(chain, wm.backEndClient, workflow.Status)
	// return executor
	return nil
}

// func createAuthorResponse(author models.DbAuthor) gen.Author {
// 	respAuthor := gen.Author{
// 		FirstName: author.FirstName,
// 		LastName:  author.LastName,
// 		Avatar: gen.Avatar{
// 			FileKey:  author.Avatar.FileKey,
// 			FileName: author.Avatar.FileName,
// 			MimeType: author.Avatar.MimeType,
// 			Size:     int(author.Avatar.Size),
// 			Url:      author.Avatar.Url,
// 		},
// 	}

// 	if author.Surname != nil {
// 		respAuthor.Surname = *author.Surname
// 	}

// 	return respAuthor
// }

// func createDbWorkflow(name string, status gen.WorkflowStatus, data json.RawMessage, authorProfileId string, createdAt, updatedAt time.Time) models.DbWorkflow {
// 	return models.DbWorkflow{
// 		Name:           name,
// 		Status:         string(status),
// 		Data:           data,
// 		CreatedAt:      createdAt,
// 		UpdatedAt:      updatedAt,
// 		AuthorProfleId: authorProfileId,
// 	}
// }

func createWorkflowData(workflow gen.Workflow) (json.RawMessage, error) {
	bytesData, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}
	return bytesData, nil
}

// func createWorkflowFromDto(dto interface{}, author gen.Author, timestamp time.Time) (gen.Workflow, error) {
// 	var workflow gen.Workflow

// 	switch d := dto.(type) {
// 	case gen.WorkflowCreateOneDto:
// 		workflow = gen.Workflow{
// 			Name:         d.Name,
// 			Status:       d.Status,
// 			CreatedAt:    timestamp,
// 			UpdatedAt:    timestamp,
// 			Nodes:        d.Nodes,
// 			Edges:        d.Edges,
// 			UpdateAuthor: author,
// 		}
// 	case gen.WorkflowUpdateOneDto:
// 		workflow = gen.Workflow{
// 			Name:         d.Name,
// 			Status:       d.Status,
// 			UpdatedAt:    timestamp,
// 			Nodes:        d.Nodes,
// 			Edges:        d.Edges,
// 			UpdateAuthor: author,
// 		}
// 	default:
// 		return gen.Workflow{}, fmt.Errorf("unsupported DTO type: %T", dto)
// 	}

// 	return workflow, nil
// }

func (wm *Manager) handleWorkflowError(id string, err error) error {
	var pgErr *pgconn.PgError
	switch {

	case errors.As(err, &pgErr) && pgErr.ConstraintName == "workflows_name_unique":
		wm.logger.Errorf("failed to save workflow %s: %v", id, err)
		return ErrorWorkflowExists

	default:
		wm.logger.Errorf("failed to save workflow %s: %v", id, err)
		return err
	}
}
