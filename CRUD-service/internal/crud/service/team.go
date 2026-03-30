package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TeamService interface {
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Create(s model.Team) (model.Team, error)
	Update(id int, s model.Team) (model.Team, error)
	Delete(id int) error
}

type teamService struct {
	db   *sqlx.DB
	repo repository.TeamRepository
}

// NewSubthemeService returns a new SubthemeService.
func NewTeamService(db *sqlx.DB, repo repository.TeamRepository) TeamService {
	return &teamService{db: db, repo: repo}
}

func (s *teamService) GetAll() ([]model.Team, error) {
	return s.repo.GetAll()
}

func (s *teamService) GetByID(id int) (model.Team, error) {
	return s.repo.GetByID(id)
}

func (s *teamService) Create(t model.Team) (model.Team, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Team{}, fmt.Errorf("teamService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	created, err := s.repo.Create(tx, t)
	if err != nil {
		return model.Team{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.Team{}, fmt.Errorf("teamService.Create commit transaction: %w", err)
	}
	return created, nil
}

func (s *teamService) Update(id int, t model.Team) (model.Team, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Team{}, fmt.Errorf("teamService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	updated, err := s.repo.Update(tx, id, t)
	if err != nil {
		return model.Team{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.Team{}, fmt.Errorf("teamService.Update commit transaction: %w", err)
	}
	return updated, nil
}

func (s *teamService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("teamService.Delete start transaction: %w", err)
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
		return fmt.Errorf("teamService.Delete commit transaction: %w", err)
	}
	return nil
}
