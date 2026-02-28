package service

import (
	"crud-service/internal/model"
	"crud-service/internal/repository"
)

// SlotService defines business-logic operations for Slot.
type SlotService interface {
	GetAll() ([]model.Slot, error)
	GetByID(id int) (model.Slot, error)
	Create(s model.Slot) (model.Slot, error)
	Update(id int, s model.Slot) (model.Slot, error)
	Delete(id int) error
}

type slotService struct {
	repo repository.SlotRepository
}

// NewSlotService returns a new SlotService.
func NewSlotService(repo repository.SlotRepository) SlotService {
	return &slotService{repo: repo}
}

func (s *slotService) GetAll() ([]model.Slot, error) {
	return s.repo.GetAll()
}

func (s *slotService) GetByID(id int) (model.Slot, error) {
	return s.repo.GetByID(id)
}

func (s *slotService) Create(sl model.Slot) (model.Slot, error) {
	return s.repo.Create(sl)
}

func (s *slotService) Update(id int, sl model.Slot) (model.Slot, error) {
	return s.repo.Update(id, sl)
}

func (s *slotService) Delete(id int) error {
	return s.repo.Delete(id)
}
