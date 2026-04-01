package service

import (
	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
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
	db          *sqlx.DB
	empoyeeRepo repository.EmployeeRepository
	slotsRepo   repository.SlotRepository
}

// NewEmployeeService returns a new EmployeeService.
func NewEmployeeService(db *sqlx.DB, empoyeeRepo repository.EmployeeRepository, slotsRepo repository.SlotRepository) EmployeeService {
	return &employeeService{db: db, empoyeeRepo: empoyeeRepo, slotsRepo: slotsRepo}
}

func (s *employeeService) GetAll() ([]model.Employee, error) {
	return s.empoyeeRepo.GetAll()
}

func (s *employeeService) GetByID(id int) (model.Employee, error) {
	return s.empoyeeRepo.GetByID(id)
}

func (s *employeeService) Create(e model.Employee) (model.Employee, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeService.Create start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	employee, err := s.empoyeeRepo.Create(tx, e)
	if err != nil {
		return model.Employee{}, err
	}

	for i := 0; i < employee.Limit; i++ {
		_, err := s.slotsRepo.Create(tx, model.Slot{EmployeeID: employee.ID, NeedToRemove: false, AppealID: nil})
		if err != nil {
			return model.Employee{}, fmt.Errorf("employeeService.Create create slot: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return model.Employee{}, fmt.Errorf("employeeService.Create commit transaction: %w", err)
	}

	return employee, nil
}

func (s *employeeService) Update(id int, e model.Employee) (model.Employee, error) {
	_, err := s.empoyeeRepo.GetByID(id)
	if err != nil {
		return model.Employee{}, fmt.Errorf("Не удалось получить сотрудника: %s", err.Error())
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return model.Employee{}, fmt.Errorf("employeeService.Update start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := s.empoyeeRepo.Update(tx, id, e)
	if err != nil {
		return model.Employee{}, fmt.Errorf("Не удалось обновить сотрудника: %s", err.Error())
	}

	if err = tx.Commit(); err != nil {
		return model.Employee{}, fmt.Errorf("employeeService.Update commit transaction: %w", err)
	}
	return updated, nil
}

func (s *employeeService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("employeeService.Delete start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.empoyeeRepo.Delete(tx, id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("employeeService.Delete commit transaction: %w", err)
	}
	return nil
}
