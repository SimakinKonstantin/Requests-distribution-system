package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
	"fmt"
	"log/slog"
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
	empoyeeRepo repository.EmployeeRepository
	slotsRepo   repository.SlotRepository
}

// NewEmployeeService returns a new EmployeeService.
func NewEmployeeService(empoyeeRepo repository.EmployeeRepository, slotsRepo repository.SlotRepository) EmployeeService {
	return &employeeService{empoyeeRepo: empoyeeRepo, slotsRepo: slotsRepo}
}

func (s *employeeService) GetAll() ([]model.Employee, error) {
	return s.empoyeeRepo.GetAll()
}

func (s *employeeService) GetByID(id int) (model.Employee, error) {
	return s.empoyeeRepo.GetByID(id)
}

func (s *employeeService) Create(e model.Employee) (model.Employee, error) {
	employee, err := s.empoyeeRepo.Create(e)
	if err != nil {
		return model.Employee{}, err
	}

	err = s.slotsRepo.CreateSlots(employee.ID, employee.Limit)
	if err != nil {
		slog.Error("Не удалось создать слоты: %s. Rollback создания сотрудника", err.Error())
		rberr := s.empoyeeRepo.Delete(employee.ID)
		if rberr != nil {
			slog.Error("Не удалось откатить создание сотрудника: %s", rberr.Error())
		}
		return model.Employee{}, fmt.Errorf("Не удалось создать слоты: %s. Rollback создания сотрудника: %s", err.Error(), rberr.Error())
	}
	return employee, nil
}

func (s *employeeService) Update(id int, e model.Employee) (model.Employee, error) {
	currentEmployee, err := s.empoyeeRepo.GetByID(id)
	if err != nil {
		return model.Employee{}, fmt.Errorf("Не удалось получить сотрудника: %s", err.Error())
	}

	updated, err := s.empoyeeRepo.Update(id, e)
	if err != nil {
		return model.Employee{}, fmt.Errorf("Не удалось обновить сотрудника: %s", err.Error())
	}

	err = s.slotsRepo.SetSlotsCount(updated.ID, updated.Limit)
	if err != nil {
		slog.Error("Не удалось обновить количество слотов: %s", err.Error())

		var rberr error
		updated, rberr = s.empoyeeRepo.Update(id, currentEmployee)
		if rberr != nil {
			slog.Error("Не удалось откатить обновление сотрудника: %s", rberr.Error())
		}
		return model.Employee{}, fmt.Errorf("Не удалось обновить количество слотов: %s. Rollback обновления сотрудника: %s", err.Error(), rberr.Error())
	}

	return updated, nil
}

func (s *employeeService) Delete(id int) error {
	return s.empoyeeRepo.Delete(id)
}
