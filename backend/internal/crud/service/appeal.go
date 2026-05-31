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

type BalancerEventPublisher interface {
	PublishAppealNeedsDistribution(appealID, teamID int) error
}

type AppealService interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(a model.Appeal) (model.Appeal, error)
	Update(id int, a model.Appeal) (model.Appeal, error)
	Delete(id int) error
	Close(id int) (model.Appeal, error)
	Assign(appealID int, employeeID int, slotID int) error
	UpsertPendingAppealByID(appealID int) error
	FetchPendingAppeals(limit int) ([]PendingAppeal, error)
	IsImportant(appealID int) bool
}

type appealService struct {
	db                *sqlx.DB
	appealRepo        repository.AppealRepository
	teamRepo          repository.TeamRepository
	clientRepo        repository.ClientRepository
	slotRepo          repository.SlotRepository
	pendingAppealRepo repository.PendingAppealRepository
	workflowService   *workflow.WorkflowService
	teamService       TeamService
	eventPublisher    BalancerEventPublisher
}

func NewAppealService(db *sqlx.DB, appealRepo repository.AppealRepository, teamRepo repository.TeamRepository, clientRepo repository.ClientRepository, slotRepo repository.SlotRepository, pendingAppealRepo repository.PendingAppealRepository, workflowService *workflow.WorkflowService, teamService TeamService, eventPublisher BalancerEventPublisher) AppealService {
	return &appealService{
		db:                db,
		appealRepo:        appealRepo,
		teamRepo:          teamRepo,
		clientRepo:        clientRepo,
		slotRepo:          slotRepo,
		pendingAppealRepo: pendingAppealRepo,
		workflowService:   workflowService,
		teamService:       teamService,
		eventPublisher:    eventPublisher,
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

	newTeam, err := s.teamService.GetTeam(a.ThemeID, a.SubthemeID, client.IsVIP)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось найти команду: %s", err.Error())
	}

	err = s.teamRepo.AssignTeam(tx, createdAppeal.ID, newTeam.ID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось назначить команду: %s", err.Error())
	}

	slog.Warn(fmt.Sprintf("BEFORE WORKFLOW: %+v", createdAppeal))

	if err = tx.Commit(); err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Create commit transaction: %w", err)
	}

	if s.workflowService != nil {
		s.workflowService.Run(map[string]interface{}{
			"appealId":                        createdAppeal.ID,
			string(workflow.ThemeId):          createdAppeal.ThemeID,
			string(workflow.Text):             createdAppeal.Text,
			string(workflow.MessageCreatedAt): time.Now().Format(time.RFC3339),
			string(workflow.ClientEmail):      client.Email,
		})
	}

	slog.Warn("WORKFLOWS FINISHED")

	if s.eventPublisher != nil {
		if pubErr := s.eventPublisher.PublishAppealNeedsDistribution(createdAppeal.ID, newTeam.ID); pubErr != nil {
			slog.Warn("appealService.Create: publish APPEAL_NEEDS_DISTRIBUTION failed", "appealId", createdAppeal.ID, "err", pubErr)
		}
	}

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

	if _, err := s.Close(id); err != nil {
		return fmt.Errorf("appealService.Delete close appeal: %w", err)
	}

	if err = s.appealRepo.Delete(tx, id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("appealService.Delete commit transaction: %w", err)
	}
	return nil
}

func (s *appealService) Close(id int) (model.Appeal, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Close start transaction: %w", err)
	}

	defer func() {
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			err = tx.Commit()
			if err != nil {
				slog.Error(fmt.Sprintf("Error committing transaction: %v", err))
			}
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	closedAppeal, err := s.appealRepo.Close(tx, id)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Close close appeal: %w", err)
	}

	slot, err := s.slotRepo.GetSlotByAppealID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return closedAppeal, nil
		}

		return model.Appeal{}, fmt.Errorf("appealService.Close get slot: %w", err)
	}

	if slot.NeedToRemove {
		err = s.slotRepo.Delete(tx, slot.ID)
		if err != nil {
			return model.Appeal{}, fmt.Errorf("appealService.Close delete slot: %w", err)
		}

		return closedAppeal, nil
	}

	needToRemoveSlot, err := s.slotRepo.GetNeedToRemoveSlot(slot.EmployeeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.Appeal{}, fmt.Errorf("appealService.Close get need to remove slot: %w", err)
	}

	// Если есть слот, помеченный на удаление, переносим с него всю информацию на этот слот.
	if !errors.Is(err, sql.ErrNoRows) {
		_, err = s.slotRepo.Update(tx, slot.ID, model.Slot{EmployeeID: needToRemoveSlot.EmployeeID, AppealID: needToRemoveSlot.AppealID, NeedToRemove: false})
		if err != nil {
			return model.Appeal{}, fmt.Errorf("appealService.Close update need to remove slot: %w", err)
		}

		err = s.slotRepo.Delete(tx, needToRemoveSlot.ID)
		if err != nil {
			return model.Appeal{}, fmt.Errorf("appealService.Close delete need to remove slot: %w", err)
		}

		return closedAppeal, nil
	}

	_, err = s.slotRepo.Update(tx, slot.ID, model.Slot{EmployeeID: slot.EmployeeID, AppealID: nil, NeedToRemove: false})
	if err != nil {
		return model.Appeal{}, fmt.Errorf("appealService.Close update slot: %w", err)
	}

	return closedAppeal, nil
}

func (s *appealService) IsImportant(appealID int) bool {
	appeal, err := s.appealRepo.GetByID(appealID)
	if err != nil {
		slog.Error(fmt.Sprintf("IsImportant: Error getting appeal: %v", err))
		return false
	}

	client, err := s.clientRepo.GetByID(appeal.ClientID)
	if err != nil {
		slog.Error(fmt.Sprintf("IsImportant: Error getting client: %v", err))
		return false
	}

	return client.IsVIP
}

type PendingAppeal struct {
	model.Appeal
	IsImportant bool `json:"isImportant"`
}

func (s *appealService) FetchPendingAppeals(limit int) ([]PendingAppeal, error) {
	appeals, err := s.appealRepo.FetchPendingAppeals(limit)
	if err != nil {
		return nil, err
	}

	pendingAppeals := make([]PendingAppeal, len(appeals))
	for i, appeal := range appeals {
		pendingAppeals[i] = PendingAppeal{Appeal: appeal, IsImportant: s.IsImportant(appeal.ID)}
	}

	return pendingAppeals, nil
}

var ErrAppealAlreadyAssigned = errors.New("appeal already assigned")
var ErrAppealClosed = errors.New("appeal closed")

func (s *appealService) Assign(appealID int, employeeID int, slotID int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("appealService.Assign start transaction: %w", err)
	}

	defer func() {
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			err = tx.Commit()
			if err != nil {
				slog.Error(fmt.Sprintf("Error committing transaction: %v", err))
			}
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	var existingEmployeeID *int
	var status string
	if err := tx.QueryRow(`SELECT employee_id, status FROM appeals WHERE id = $1 FOR UPDATE`, appealID).Scan(&existingEmployeeID, &status); err != nil {
		return fmt.Errorf("appealService.Assign get appeal: %w", err)
	}

	if status == "closed" {
		return fmt.Errorf("appealService.Assign appeal %d is closed: %w", appealID, ErrAppealClosed)
	}

	if existingEmployeeID != nil {
		return ErrAppealAlreadyAssigned
	}

	var slotAppealID *int
	if err := tx.QueryRow(`SELECT appeal_id FROM slots WHERE id = $1 AND employee_id = $2 FOR UPDATE`, slotID, employeeID).Scan(&slotAppealID); err != nil {
		return fmt.Errorf("appealService.Assign get slot: %w", err)
	}
	// Защита от гонки: если слот уже занят другим обращением, не делаем назначение повторно.
	if slotAppealID != nil {
		return fmt.Errorf("appealService.Assign slot %d already occupied: %w", slotID, sql.ErrNoRows)
	}

	_, err = tx.Exec(`UPDATE appeals SET employee_id = $1, status = 'active' WHERE id = $2`, employeeID, appealID)
	if err != nil {
		return fmt.Errorf("appealService.Assign update appeal: %w", err)
	}

	_, err = tx.Exec(`UPDATE slots SET appeal_id = $1, updated_at = now() WHERE id = $2`, appealID, slotID)
	if err != nil {
		return fmt.Errorf("appealService.Assign update slot: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM pending_appeals WHERE appeal_id = $1`, appealID)
	if err != nil {
		return fmt.Errorf("appealService.Assign delete pending appeal: %w", err)
	}

	_, err = tx.Exec(`UPDATE employees SET last_assign_at = now() WHERE id = $1`, employeeID)
	if err != nil {
		return fmt.Errorf("appealService.Assign update employee: %w", err)
	}

	return nil
}

func (s *appealService) UpsertPendingAppealByID(appealID int) error {
	appealInfo, err := s.appealRepo.GetByID(appealID)
	if err != nil {
		return fmt.Errorf("appealService.UpsertPendingAppealByID get appeal: %w", err)
	}

	if appealInfo.Status == "closed" {
		return s.pendingAppealRepo.RemovePendingAppeal(appealID)
	}
	if appealInfo.EmployeeID != nil {
		return s.pendingAppealRepo.RemovePendingAppeal(appealID)
	}
	priority := calculateAppealPriority(s.IsImportant(appealID), appealInfo.CreatedAt)

	err = s.pendingAppealRepo.UpsertPendingAppealByID(appealID, priority)
	if err != nil {
		return fmt.Errorf("appealService.UpsertPendingAppealByID upsert pending appeal: %w", err)
	}

	return err
}

// Функция вычисления приоритета обращения.
func calculateAppealPriority(isImportant bool, createdAt time.Time) int64 {
	now := time.Now()

	// Базовый приоритет рассчитывается в зависимости от статуса пользователя.
	var base int64
	if isImportant {
		base = 2

	} else {
		base = 1
	}

	priority := int64(base * 100)

	// Рассчет времени жизни обращения в минутах.
	ageMinutes := int64(now.Sub(createdAt).Minutes())
	priority += ageMinutes * 10

	return priority
}
