package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
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
	db   *sqlx.DB
	repo repository.SubthemeRepository
}

// NewSubthemeService returns a new SubthemeService.
func NewSubthemeService(db *sqlx.DB, repo repository.SubthemeRepository) SubthemeService {
	return &subthemeService{db: db, repo: repo}
}

func (s *subthemeService) GetAll() ([]model.Subtheme, error) {
	return s.repo.GetAll()
}

func (s *subthemeService) GetByID(id int) (model.Subtheme, error) {
	return s.repo.GetByID(id)
}

func (s *subthemeService) Create(st model.Subtheme) (model.Subtheme, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	created, err := s.repo.Create(tx, st)
	if err != nil {
		return model.Subtheme{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeService.Create commit transaction: %w", err)
	}
	return created, nil
}

func (s *subthemeService) Update(id int, st model.Subtheme) (model.Subtheme, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := s.repo.Update(tx, id, st)
	if err != nil {
		return model.Subtheme{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Subtheme{}, fmt.Errorf("subthemeService.Update commit transaction: %w", err)
	}
	return updated, nil
}

func (s *subthemeService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("subthemeService.Delete start transaction: %w", err)
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
		return fmt.Errorf("subthemeService.Delete commit transaction: %w", err)
	}
	return nil
}
