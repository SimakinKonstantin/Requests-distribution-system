package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

// ThemeService defines business-logic operations for Theme.
type ThemeService interface {
	GetAll() ([]model.Theme, error)
	GetByID(id int) (model.Theme, error)
	Create(t model.Theme) (model.Theme, error)
	Update(id int, t model.Theme) (model.Theme, error)
	Delete(id int) error
}

type themeService struct {
	repo repository.ThemeRepository
}

// NewThemeService returns a new ThemeService.
func NewThemeService(repo repository.ThemeRepository) ThemeService {
	return &themeService{repo: repo}
}

func (s *themeService) GetAll() ([]model.Theme, error)                  { return s.repo.GetAll() }
func (s *themeService) GetByID(id int) (model.Theme, error)             { return s.repo.GetByID(id) }
func (s *themeService) Create(t model.Theme) (model.Theme, error)       { return s.repo.Create(t) }
func (s *themeService) Update(id int, t model.Theme) (model.Theme, error) {
	return s.repo.Update(id, t)
}
func (s *themeService) Delete(id int) error { return s.repo.Delete(id) }
