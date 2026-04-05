package service

import (
	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TeamService interface {
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Create(s model.Team) (model.Team, error)
	Update(id int, s model.Team) (model.Team, error)
	Delete(id int) error
	AssignTeam(appealId int, teamId int) error
	GetTeam(themeID int, subthemeID *int, isVIP bool) (team model.Team, err error)
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

func (s *teamService) AssignTeam(appealId int, teamId int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("teamService.AssignTeam start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	err = s.repo.AssignTeam(tx, appealId, teamId)
	if err != nil {
		return fmt.Errorf("teamService.AssignTeam: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("teamService.AssignTeam commit transaction: %w", err)
	}
	return nil
}

const DEFAULT_TEAM_NAME_NOT_VIP = "Не распределенные"
const DEFAULT_TEAM_NAME_VIP = "Не распределенные VIP"

func (s *teamService) GetTeam(themeID int, subthemeID *int, isVIP bool) (team model.Team, err error) {
	// Пробуем получить команду.
	team, err = s.repo.GetTeamByThemeSubtheme(themeID, subthemeID, isVIP)
	if errors.Is(err, sql.ErrNoRows) {

		// Если такой команды нет, то пробуем получить игнорируя подтему.
		team, err = s.repo.GetTeamByThemeSubtheme(themeID, nil, isVIP)
		if errors.Is(err, sql.ErrNoRows) {
			if isVIP {
				team, err = s.repo.GetTeamByName(DEFAULT_TEAM_NAME_VIP)

			} else {
				team, err = s.repo.GetTeamByName(DEFAULT_TEAM_NAME_NOT_VIP)
			}
		}

		if err != nil {
			return team, fmt.Errorf("teamService.GetTeam error finding team: %w", err)
		}
	}
	if err != nil {
		return team, fmt.Errorf("teamService.GetTeam error finding team: %w", err)
	}

	return
}
