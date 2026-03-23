package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

type TeamService interface {
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Create(s model.Team) (model.Team, error)
	Update(id int, s model.Team) (model.Team, error)
	Delete(id int) error
}

type teamService struct {
	repo repository.TeamRepository
}

// NewSubthemeService returns a new SubthemeService.
func NewTeamService(repo repository.TeamRepository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) GetAll() ([]model.Team, error) {
	return s.repo.GetAll()
}

func (s *teamService) GetByID(id int) (model.Team, error) {
	return s.repo.GetByID(id)
}

func (s *teamService) Create(t model.Team) (model.Team, error) {
	return s.repo.Create(t)
}

func (s *teamService) Update(id int, t model.Team) (model.Team, error) {
	return s.repo.Update(id, t)
}

func (s *teamService) Delete(id int) error {
	return s.repo.Delete(id)
}
