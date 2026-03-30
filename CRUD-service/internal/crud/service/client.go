package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ClientService defines business-logic operations for Client.
type ClientService interface {
	GetAll() ([]model.Client, error)
	GetByID(id int) (model.Client, error)
	Create(c model.Client) (model.Client, error)
	Update(id int, c model.Client) (model.Client, error)
	Delete(id int) error
}

type clientService struct {
	db   *sqlx.DB
	repo repository.ClientRepository
}

// NewClientService returns a new ClientService.
func NewClientService(db *sqlx.DB, repo repository.ClientRepository) ClientService {
	return &clientService{db: db, repo: repo}
}

func (s *clientService) GetAll() ([]model.Client, error) {
	return s.repo.GetAll()
}

func (s *clientService) GetByID(id int) (model.Client, error) {
	return s.repo.GetByID(id)
}

func (s *clientService) Create(c model.Client) (model.Client, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Client{}, fmt.Errorf("clientService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	created, err := s.repo.Create(tx, c)
	if err != nil {
		return model.Client{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Client{}, fmt.Errorf("clientService.Create commit transaction: %w", err)
	}
	return created, nil
}

func (s *clientService) Update(id int, c model.Client) (model.Client, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Client{}, fmt.Errorf("clientService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := s.repo.Update(tx, id, c)
	if err != nil {
		return model.Client{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Client{}, fmt.Errorf("clientService.Update commit transaction: %w", err)
	}
	return updated, nil
}

func (s *clientService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("clientService.Delete start transaction: %w", err)
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
		return fmt.Errorf("clientService.Delete commit transaction: %w", err)
	}
	return nil
}
