package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

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
	repo repository.AppealRepository
}

// NewAppealService returns a new AppealService.
func NewAppealService(repo repository.AppealRepository) AppealService {
	return &appealService{repo: repo}
}

func (s *appealService) GetAll() ([]model.Appeal, error) {
	return s.repo.GetAll()
}

func (s *appealService) GetByID(id int) (model.Appeal, error) {
	return s.repo.GetByID(id)
}

func (s *appealService) Create(a model.Appeal) (model.Appeal, error) {
	return s.repo.Create(a)
}

func (s *appealService) Update(id int, a model.Appeal) (model.Appeal, error) {
	return s.repo.Update(id, a)
}

func (s *appealService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *appealService) Close(id int) (model.Appeal, error) {
	return s.repo.Close(id)
}
