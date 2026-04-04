package service

import (
	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"crud-service/internal/workflow"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

const DEFAULT_TEAM_NAME = "Не распределенные"

// AppealService defines business-logic operations for Appeal.
type AppealService interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(a model.Appeal) (model.Appeal, error)
	Update(id int, a model.Appeal) (model.Appeal, error)
	Delete(id int) error
	Close(id int) error
}

type appealService struct {
	db              *sqlx.DB
	appealRepo      repository.AppealRepository
	teamRepo        repository.TeamRepository
	clientRepo      repository.ClientRepository
	slotRepo        repository.SlotRepository
	workflowService workflow.WorkflowService
}

// NewAppealService returns a new AppealService.
func NewAppealService(db *sqlx.DB, appealRepo repository.AppealRepository, teamRepo repository.TeamRepository, clientRepo repository.ClientRepository, slotRepo repository.SlotRepository, workflowService workflow.WorkflowService) AppealService {
	return &appealService{
		db:              db,
		appealRepo:      appealRepo,
		teamRepo:        teamRepo,
		clientRepo:      clientRepo,
		slotRepo:        slotRepo,
		workflowService: workflowService,
	}
}

func (s *appealService) GetAll() ([]model.Appeal, error) {
	return s.appealRepo.GetAll()
}

func (s *appealService) GetByID(id int) (model.Appeal, error) {
	return s.appealRepo.GetByID(id)
}

func (s *appealService) Create(a model.Appeal) (model.Appeal, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	createdAppeal, err := s.appealRepo.Create(tx, a)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось создать обращение: %s", err.Error())
	}

	client, err := s.clientRepo.GetByID(createdAppeal.ClientID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось найти клиента: %s", err.Error())
	}

	_, err = s.teamRepo.GetTeamByThemeSubtheme(createdAppeal.ThemeID, createdAppeal.SubthemeID, client.IsVIP)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.Appeal{}, fmt.Errorf("Не удалось найти команду: %s", err.Error())
	}

	if errors.Is(err, sql.ErrNoRows) {
		defaultTeam, err := s.teamRepo.GetTeamByName(DEFAULT_TEAM_NAME)
		if err != nil {
			return model.Appeal{}, fmt.Errorf("Не удалось найти команду: %s", err.Error())
		}
		createdAppeal.TeamID = &defaultTeam.ID
	}

	slog.Warn(fmt.Sprintf("BEFORE WORKFLOW: %+v", createdAppeal))

	if err = tx.Commit(); err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Create commit transaction: %w", err)
	}

	s.workflowService.Run(map[string]interface{}{
		"appealId":                        createdAppeal.ID,
		string(workflow.ThemeId):          createdAppeal.ThemeID,
		string(workflow.Text):             createdAppeal.Text,
		string(workflow.MessageCreatedAt): time.Now().Format(time.RFC3339),
		string(workflow.ClientEmail):      client.Email,
	})

	slog.Warn("WORKFLOWS FINISHED")

	return createdAppeal, nil
}

func (s *appealService) Update(id int, a model.Appeal) (model.Appeal, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := s.appealRepo.Update(tx, id, a)
	if err != nil {
		return model.Appeal{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Update commit transaction: %w", err)
	}
	return updated, nil
}

func (s *appealService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("appealService.Delete start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.appealRepo.Delete(tx, id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("appealService.Delete commit transaction: %w", err)
	}
	return nil
}

func (s *appealService) Close(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("appealService.Close start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = s.appealRepo.Close(tx, id)
	if err != nil {
		return fmt.Errorf("appealService.Close close appeal: %w", err)
	}

	slot, err := s.slotRepo.GetSlotByAppealID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("appealService.Close get slot: %w", err)
	}

	if slot.NeedToRemove {
		err = s.slotRepo.Delete(tx, slot.ID)
		if err != nil {
			return fmt.Errorf("appealService.Close delete slot: %w", err)
		}

		return nil
	}

	needToRemoveSlot, err := s.slotRepo.GetNeedToRemoveSlot(slot.EmployeeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("appealService.Close get need to remove slot: %w", err)
	}

	// Если есть needToRemoveSlot, то всю информацию с него переносим на этот слот.
	if !errors.Is(err, sql.ErrNoRows) {
		_, err = s.slotRepo.Update(tx, slot.ID, model.Slot{EmployeeID: needToRemoveSlot.EmployeeID, AppealID: needToRemoveSlot.AppealID, NeedToRemove: false})
		if err != nil {
			return fmt.Errorf("appealService.Close update need to remove slot: %w", err)
		}

		err = s.slotRepo.Delete(tx, needToRemoveSlot.ID)
		if err != nil {
			return fmt.Errorf("appealService.Close delete need to remove slot: %w", err)
		}

		return nil
	}

	_, err = s.slotRepo.Update(tx, slot.ID, model.Slot{EmployeeID: slot.EmployeeID, AppealID: nil, NeedToRemove: false})
	if err != nil {
		return fmt.Errorf("appealService.Close update slot: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("appealService.Close commit transaction: %w", err)
	}

	return nil
}
