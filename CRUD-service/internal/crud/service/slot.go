package service

import (
	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SlotService defines business-logic operations for Slot.
type SlotService interface {
	GetAll() ([]model.Slot, error)
	GetByID(id int) (model.Slot, error)
	Create(s model.Slot) (model.Slot, error)
	UpdateCount(id int, count int) error
	Delete(id int) error
}

type slotService struct {
	db   *sqlx.DB
	repo repository.SlotRepository
}

func NewSlotService(db *sqlx.DB, repo repository.SlotRepository) SlotService {
	return &slotService{db: db, repo: repo}
}

func (s *slotService) GetAll() ([]model.Slot, error) {
	return s.repo.GetAll()
}

func (s *slotService) GetByID(id int) (model.Slot, error) {
	return s.repo.GetByID(id)
}

func (s *slotService) Create(slot model.Slot) (model.Slot, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotService.Create start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	createdSlot, err := s.repo.Create(tx, slot)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotService.Create create slot: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return model.Slot{}, fmt.Errorf("slotService.UpdateCount commit transaction: %w", err)
	}
	return createdSlot, nil
}

func (s *slotService) UpdateCount(employeeID int, newCount int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("slotService.Update start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	currentCount, err := s.repo.GetSlotsCount(employeeID)
	if err != nil {
		return fmt.Errorf("slotService.UpdateCount get slots count: %w", err)
	}

	if currentCount == newCount {
		return nil
	}

	if currentCount < newCount {
		needToRemoveSlots, err := s.repo.GetNeedToRemoveSlots(employeeID)
		if err != nil {
			return fmt.Errorf("slotService.UpdateCount get need to remove slots: %w", err)
		}

		for _, slot := range needToRemoveSlots {
			err := s.repo.SetNeedToRemoveValue(tx, slot, false)
			if err != nil {
				return fmt.Errorf("slotService.UpdateCount set need to remove value: %w", err)
			}

			// Если теперь слотов хватает, то выходим не добавляя новых слотов.
			currentCount++
			if currentCount == newCount {
				return nil
			}
		}

		needToAddSlots := newCount - currentCount
		for i := 0; i < needToAddSlots; i++ {
			_, err := s.repo.Create(tx, model.Slot{EmployeeID: employeeID, NeedToRemove: false, AppealID: nil})
			if err != nil {
				return fmt.Errorf("slotService.UpdateCount create slot: %w", err)
			}
		}

		return nil
	}

	if currentCount > newCount {
		freeSlots, err := s.repo.GetFreeSlots(employeeID)
		if err != nil {
			return fmt.Errorf("slotService.UpdateCount get free slots: %w", err)
		}

		// Сначала удаляем все свободные слоты.
		for _, slot := range freeSlots {
			err := s.repo.Delete(tx, slot.ID)
			if err != nil {
				return fmt.Errorf("slotService.UpdateCount delete slot: %w", err)
			}
			currentCount--
			if currentCount == newCount {
				return nil
			}
		}

		// Если после удаления свободных слотов
		needToRemoveSlots := currentCount - newCount
		for i := 0; i < needToRemoveSlots; i++ {
			err := s.repo.SetNeedToRemoveValue(tx, freeSlots[i], true)
			if err != nil {
				return fmt.Errorf("slotService.UpdateCount set need to remove value: %w", err)
			}
		}
		return nil
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("slotService.UpdateCount commit transaction: %w", err)
	}
	return nil
}

func (s *slotService) Delete(id int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("slotService.Delete start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = s.repo.Delete(tx, id)
	if err != nil {
		return fmt.Errorf("slotService.Delete delete slot: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("slotService.Close commit transaction: %w", err)
	}
	return nil
}

func (s *slotService) FetchFreeSlotsByManagers(managerIDs []int) (map[int][]model.Slot, error) {
	if len(managerIDs) == 0 {
		return map[int][]model.Slot{}, nil
	}
	const q = `
SELECT id, employee_id, appeal_id, updated_at
FROM slots
WHERE employee_id = ANY($1) AND appeal_id IS NULL
ORDER BY employee_id, updated_at ASC
`
	rows, err := db.Pool.Query(ctx, q, managerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int][]SlotRow, len(managerIDs))
	for rows.Next() {
		var s SlotRow
		if err := rows.Scan(&s.ID, &s.ManagerID, &s.AppealID, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out[s.ManagerID] = append(out[s.ManagerID], s)
	}
	return out, rows.Err()
}
