package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
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
	db   *sqlx.DB
	repo repository.ThemeRepository
}

// NewThemeService returns a new ThemeService.
func NewThemeService(db *sqlx.DB, repo repository.ThemeRepository) ThemeService {
	return &themeService{db: db, repo: repo}
}

func (s *themeService) GetAll() ([]model.Theme, error)          { return s.repo.GetAll() }
func (s *themeService) GetByID(id int) (model.Theme, error)     { return s.repo.GetByID(id) }
func (s *themeService) Create(t model.Theme) (model.Theme, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Theme{}, fmt.Errorf("themeService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	created, err := s.repo.Create(tx, t)
	if err != nil {
		return model.Theme{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.Theme{}, fmt.Errorf("themeService.Create commit transaction: %w", err)
	}
	return created, nil
}
func (s *themeService) Update(id int, t model.Theme) (model.Theme, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Theme{}, fmt.Errorf("themeService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	updated, err := s.repo.Update(tx, id, t)
	if err != nil {
		return model.Theme{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.Theme{}, fmt.Errorf("themeService.Update commit transaction: %w", err)
	}
	return updated, nil
}
func (s *themeService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("themeService.Delete start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = s.repo.Delete(tx, id); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("themeService.Delete commit transaction: %w", err)
	}
	return nil
}
