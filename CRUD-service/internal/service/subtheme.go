package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

// SubthemeService defines business-logic operations for Subtheme.
type SubthemeService interface {
	GetAll() ([]model.Subtheme, error)
	GetByID(id int) (model.Subtheme, error)
	Create(s model.Subtheme) (model.Subtheme, error)
	Update(id int, s model.Subtheme) (model.Subtheme, error)
	Delete(id int) error
}

type subthemeService struct {
	repo repository.SubthemeRepository
}

// NewSubthemeService returns a new SubthemeService.
func NewSubthemeService(repo repository.SubthemeRepository) SubthemeService {
	return &subthemeService{repo: repo}
}

func (s *subthemeService) GetAll() ([]model.Subtheme, error) {
	return s.repo.GetAll()
}

func (s *subthemeService) GetByID(id int) (model.Subtheme, error) {
	return s.repo.GetByID(id)
}

func (s *subthemeService) Create(st model.Subtheme) (model.Subtheme, error) {
	return s.repo.Create(st)
}

func (s *subthemeService) Update(id int, st model.Subtheme) (model.Subtheme, error) {
	return s.repo.Update(id, st)
}

func (s *subthemeService) Delete(id int) error {
	return s.repo.Delete(id)
}
