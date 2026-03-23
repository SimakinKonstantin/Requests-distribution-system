package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"
)

const DEFAULT_TEAM_NAME = "Не распределенные"

// AppealService defines business-logic operations for Appeal.
type AppealService interface {
	GetAll() ([]model.Appeal, error)
	GetByID(id int) (model.Appeal, error)
	Create(a model.Appeal) (model.Appeal, error)
	Update(id int, a model.Appeal) (model.Appeal, error)
	Delete(id int) error
	Close(id int) (model.Appeal, error)
}

type appealService struct {
	appealRepo repository.AppealRepository
	teamRepo   repository.TeamRepository
	clientRepo repository.ClientRepository
}

// NewAppealService returns a new AppealService.
func NewAppealService(repo repository.AppealRepository) AppealService {
	return &appealService{appealRepo: repo}
}

func (s *appealService) GetAll() ([]model.Appeal, error) {
	return s.appealRepo.GetAll()
}

func (s *appealService) GetByID(id int) (model.Appeal, error) {
	return s.appealRepo.GetByID(id)
}

func (s *appealService) Create(a model.Appeal) (model.Appeal, error) {
	createdAppeal, err := s.appealRepo.Create(a)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось создать обращение: %s", err.Error())
	}

	client, err := s.clientRepo.GetByID(createdAppeal.ClientID)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось найти клиента: %s", err.Error())
	}

	team, err := s.teamRepo.GetTeamByThemeSubtheme(createdAppeal.ThemeID, createdAppeal.SubthemeID, client.IsVIP)
	if err != nil {
		return model.Appeal{}, fmt.Errorf("Не удалось найти команду: %s", err.Error())
	}

	if team.ID == 0 {
		defaultTeam, err := s.teamRepo.GetTeamByName(DEFAULT_TEAM_NAME)
		if err != nil {
			return model.Appeal{}, fmt.Errorf("Не удалось найти команду: %s", err.Error())
		}
		createdAppeal.TeamID = &defaultTeam.ID
	}

	// Запускаем workflow

	return createdAppeal, nil
}

func (s *appealService) Update(id int, a model.Appeal) (model.Appeal, error) {
	return s.appealRepo.Update(id, a)
}

func (s *appealService) Delete(id int) error {
	return s.appealRepo.Delete(id)
}

func (s *appealService) Close(id int) (model.Appeal, error) {
	return s.appealRepo.Close(id)
}
