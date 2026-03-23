package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// slotDB is the database-level representation of Slot.
type slotDB struct {
	ID           int  `db:"id"`
	EmployeeID   int  `db:"employee_id"`
	AppealID     *int `db:"appeal_id"`
	NeedToRemove bool `db:"need_to_remove"`
}

func toSlotDB(s model.Slot) slotDB {
	return slotDB{
		ID:         s.ID,
		EmployeeID: s.EmployeeID,
		AppealID:   s.AppealID,
	}
}

func (s slotDB) toDomain() model.Slot {
	return model.Slot{
		ID:         s.ID,
		EmployeeID: s.EmployeeID,
		AppealID:   s.AppealID,
	}
}

// SlotRepository defines CRUD operations for Slot.
type SlotRepository interface {
	GetAll() ([]model.Slot, error)
	GetByID(id int) (model.Slot, error)
	Create(s model.Slot) (model.Slot, error)
	Update(id int, s model.Slot) (model.Slot, error)
	Delete(id int) error
	CreateSlots(employeeID int, limit int) error
	SetNeedToRemoveValue(slot model.Slot, value bool) error
	GetFreeSlots(employeeID int) ([]model.Slot, error)
	GetRealSlots(employeeID int) ([]model.Slot, error)
	GetNeedToRemoveSlots(employeeID int) ([]model.Slot, error)
	GetSlotsCount(employeeID int) (int, error)
	SetSlotsCount(employeeID int, count int) error
}

type slotRepo struct {
	db *sqlx.DB
}

// NewSlotRepository returns a PostgreSQL-backed SlotRepository.
func NewSlotRepository(db *sqlx.DB) SlotRepository {
	return &slotRepo{db: db}
}

func (r *slotRepo) GetAll() ([]model.Slot, error) {
	var rows []slotDB
	if err := r.db.Select(&rows, `SELECT id, employee_id, appeal_id FROM slots`); err != nil {
		return nil, fmt.Errorf("slotRepo.GetAll: %w", err)
	}

	result := make([]model.Slot, len(rows))
	for i, row := range rows {
		result[i] = row.toDomain()
	}
	return result, nil
}

func (r *slotRepo) GetByID(id int) (model.Slot, error) {
	var row slotDB
	err := r.db.Get(&row, `SELECT id, employee_id, appeal_id FROM slots WHERE id = $1`, id)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.GetByID: %w", wrapNotFound(err))
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Create(s model.Slot) (model.Slot, error) {
	row := toSlotDB(s)
	err := r.db.QueryRowx(
		`INSERT INTO slots (employee_id, appeal_id) VALUES ($1, $2) RETURNING id`,
		row.EmployeeID, row.AppealID,
	).Scan(&row.ID)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Update(id int, s model.Slot) (model.Slot, error) {
	row := toSlotDB(s)
	row.ID = id
	res, err := r.db.Exec(
		`UPDATE slots SET employee_id=$1, appeal_id=$2 WHERE id=$3`,
		row.EmployeeID, row.AppealID, row.ID,
	)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Slot{}, err
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM slots WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("slotRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

func (r *slotRepo) CreateSlots(employeeID int, limit int) error {
	for i := 0; i < limit; i++ {
		_, err := r.Create(model.Slot{EmployeeID: employeeID})
		if err != nil {
			return fmt.Errorf("slotRepo.CreateSlots: %w", err)
		}
	}
	return nil
}

func (r *slotRepo) SetSlotsCount(employeeID int, count int) error {
	//todo: Параллельно с обработкой, может прилететь другой запрос на изменение количества слотов. Поэтому нужно всю эту логику выполнять в транзакции.

	// Определяем Общее количество слотов
	// Определяем количество needToRemove слотов
	currentCount, err := r.GetSlotsCount(employeeID)
	if err != nil {
		return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
	}

	if currentCount == count {
		return nil
	}

	if currentCount < count {
		needToRemoveSlots, err := r.GetNeedToRemoveSlots(employeeID)
		if err != nil {
			return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
		}

		for _, slot := range needToRemoveSlots {
			err := r.SetNeedToRemoveValue(slot, false)
			if err != nil {
				return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
			}

			// Если теперь слотов хватает, то выходим не добавляя новых слотов.
			currentCount++
			if currentCount == count {
				return nil
			}
		}

		needToAddSlots := count - currentCount
		for i := 0; i < needToAddSlots; i++ {
			_, err := r.Create(model.Slot{EmployeeID: employeeID, NeedToRemove: false, AppealID: nil})
			if err != nil {
				return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
			}
		}

		return nil
	}

	if currentCount > count {
		freeSlots, err := r.GetFreeSlots(employeeID)
		if err != nil {
			return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
		}

		// Сначала удаляем все свободные слоты.
		for _, slot := range freeSlots {
			err := r.Delete(slot.ID)
			if err != nil {
				return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
			}
			currentCount--
			if currentCount == count {
				return nil
			}
		}

		// Если после удаления свободных слотов
		needToRemoveSlots := currentCount - count
		for i := 0; i < needToRemoveSlots; i++ {
			r.SetNeedToRemoveValue(freeSlots[i], true)
			if err != nil {
				return fmt.Errorf("slotRepo.SetSlotsCount: %w", err)
			}
		}
		return nil
	}

	return nil
}

func (r *slotRepo) GetSlotsCount(employeeID int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM slots WHERE employee_id = $1`, employeeID)
	if err != nil {
		return 0, fmt.Errorf("slotRepo.GetSlotsCount: %w", err)
	}
	return count, nil
}

func (r *slotRepo) GetNeedToRemoveSlots(employeeID int) ([]model.Slot, error) {
	var slots []model.Slot
	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id FROM slots WHERE employee_id = $1 AND need_to_remove = TRUE`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetNeedToRemoveSlots: %w", err)
	}
	return slots, nil
}

func (r *slotRepo) GetRealSlots(employeeID int) ([]model.Slot, error) {
	var slots []model.Slot
	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id FROM slots WHERE employee_id = $1 AND need_to_remove = FALSE`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetRealSlots: %w", err)
	}
	return slots, nil
}

func (r *slotRepo) GetFreeSlots(employeeID int) ([]model.Slot, error) {
	var slots []model.Slot
	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id FROM slots WHERE employee_id = $1 AND appeal_id IS NULL`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetRealSlots: %w", err)
	}
	return slots, nil
}

func (r *slotRepo) SetNeedToRemoveValue(slot model.Slot, value bool) error {
	_, err := r.db.Exec(`UPDATE slots SET need_to_remove = $1 WHERE id = $2`, value, slot.ID)
	if err != nil {
		return fmt.Errorf("slotRepo.SetNeedToRemoveValue: %w", err)
	}
	return nil
}
