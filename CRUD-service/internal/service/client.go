package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
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
	repo repository.ClientRepository
}

// NewClientService returns a new ClientService.
func NewClientService(repo repository.ClientRepository) ClientService {
	return &clientService{repo: repo}
}

func (s *clientService) GetAll() ([]model.Client, error)                   { return s.repo.GetAll() }
func (s *clientService) GetByID(id int) (model.Client, error)              { return s.repo.GetByID(id) }
func (s *clientService) Create(c model.Client) (model.Client, error)       { return s.repo.Create(c) }
func (s *clientService) Update(id int, c model.Client) (model.Client, error) {
	return s.repo.Update(id, c)
}
func (s *clientService) Delete(id int) error { return s.repo.Delete(id) }
