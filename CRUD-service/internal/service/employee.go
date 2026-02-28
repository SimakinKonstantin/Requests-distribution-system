package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

// EmployeeService defines business-logic operations for Employee.
type EmployeeService interface {
	GetAll() ([]model.Employee, error)
	GetByID(id int) (model.Employee, error)
	Create(e model.Employee) (model.Employee, error)
	Update(id int, e model.Employee) (model.Employee, error)
	Delete(id int) error
}

type employeeService struct {
	repo repository.EmployeeRepository
}

// NewEmployeeService returns a new EmployeeService.
func NewEmployeeService(repo repository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) GetAll() ([]model.Employee, error) {
	return s.repo.GetAll()
}

func (s *employeeService) GetByID(id int) (model.Employee, error) {
	return s.repo.GetByID(id)
}

func (s *employeeService) Create(e model.Employee) (model.Employee, error) {
	return s.repo.Create(e)
}

func (s *employeeService) Update(id int, e model.Employee) (model.Employee, error) {
	return s.repo.Update(id, e)
}

func (s *employeeService) Delete(id int) error {
	return s.repo.Delete(id)
}
