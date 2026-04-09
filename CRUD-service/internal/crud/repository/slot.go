package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/crud/model"
)

// slotDB is the database-level representation of Slot.
type slotDB struct {
	ID           int        `db:"id"`
	EmployeeID   int        `db:"employee_id"`
	AppealID     *int       `db:"appeal_id"`
	NeedToRemove bool       `db:"need_to_remove"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

func toSlotDB(s model.Slot) slotDB {
	return slotDB{
		ID:           s.ID,
		EmployeeID:   s.EmployeeID,
		AppealID:     s.AppealID,
		NeedToRemove: s.NeedToRemove,
		UpdatedAt:    s.UpdatedAt,
	}
}

func (s slotDB) toDomain() model.Slot {
	return model.Slot{
		ID:           s.ID,
		EmployeeID:   s.EmployeeID,
		AppealID:     s.AppealID,
		NeedToRemove: s.NeedToRemove,
		UpdatedAt:    s.UpdatedAt,
	}
}

// SlotRepository defines CRUD operations for Slot.
type SlotRepository interface {
	GetAll() ([]model.Slot, error)
	GetByID(id int) (model.Slot, error)
	Create(tx *sqlx.Tx, s model.Slot) (model.Slot, error)
	Update(tx *sqlx.Tx, id int, s model.Slot) (model.Slot, error)
	Delete(tx *sqlx.Tx, id int) error
	SetNeedToRemoveValue(tx *sqlx.Tx, slot model.Slot, value bool) error
	GetFreeSlots(employeeID int) ([]model.Slot, error)
	GetRealSlots(employeeID int) ([]model.Slot, error)
	GetNeedToRemoveSlots(employeeID int) ([]model.Slot, error)
	GetSlotsCount(employeeID int) (int, error)
	GetSlotByAppealID(appealID int) (model.Slot, error)
	GetNeedToRemoveSlot(employeeID int) (model.Slot, error)
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
	if err := r.db.Select(&rows, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots`); err != nil {
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
	err := r.db.Get(&row, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE id = $1`, id)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.GetByID: %w", err)
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Create(tx *sqlx.Tx, s model.Slot) (model.Slot, error) {
	row := toSlotDB(s)
	err := tx.QueryRowx(
		`INSERT INTO slots (employee_id, appeal_id, need_to_remove, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		row.EmployeeID, row.AppealID,
		row.NeedToRemove, row.UpdatedAt,
	).Scan(&row.ID)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.Create: %w", err)
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Update(tx *sqlx.Tx, id int, s model.Slot) (model.Slot, error) {
	row := toSlotDB(s)
	row.ID = id
	res, err := tx.Exec(
		`UPDATE slots SET employee_id=$1, appeal_id=$2, need_to_remove=$3, updated_at=$4 WHERE id=$5`,
		row.EmployeeID, row.AppealID,
		row.NeedToRemove, row.UpdatedAt, row.ID,
	)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.Update: %w", err)
	}
	if err = expectOneRow(res); err != nil {
		return model.Slot{}, err
	}
	return row.toDomain(), nil
}

func (r *slotRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM slots WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("slotRepo.Delete: %w", err)
	}
	return expectOneRow(res)
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
	var slots []slotDB

	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE employee_id = $1 AND need_to_remove = TRUE`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetNeedToRemoveSlots: %w", err)
	}

	result := make([]model.Slot, len(slots))
	for i, slot := range slots {
		result[i] = slot.toDomain()
	}
	return result, nil
}

func (r *slotRepo) GetRealSlots(employeeID int) ([]model.Slot, error) {
	var slots []slotDB

	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE employee_id = $1 AND need_to_remove = FALSE`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetRealSlots: %w", err)
	}

	result := make([]model.Slot, len(slots))
	for i, slot := range slots {
		result[i] = slot.toDomain()
	}
	return result, nil
}

func (r *slotRepo) GetFreeSlots(employeeID int) ([]model.Slot, error) {
	var slots []slotDB

	err := r.db.Select(&slots, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE employee_id = $1 AND appeal_id IS NULL`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("slotRepo.GetRealSlots: %w", err)
	}

	result := make([]model.Slot, len(slots))
	for i, slot := range slots {
		result[i] = slot.toDomain()
	}
	return result, nil
}

func (r *slotRepo) SetNeedToRemoveValue(tx *sqlx.Tx, slot model.Slot, value bool) error {
	_, err := tx.Exec(`UPDATE slots SET need_to_remove = $1 WHERE id = $2`, value, slot.ID)
	if err != nil {
		return fmt.Errorf("slotRepo.SetNeedToRemoveValue: %w", err)
	}
	return nil
}

func (r *slotRepo) GetSlotByAppealID(appealID int) (model.Slot, error) {
	var slot slotDB
	err := r.db.Get(&slot, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE appeal_id = $1`, appealID)

	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.GetSlotByAppealID: %w", err)
	}
	return slot.toDomain(), nil
}

func (r *slotRepo) GetNeedToRemoveSlot(employeeID int) (model.Slot, error) {
	var slot slotDB

	err := r.db.Get(&slot, `SELECT id, employee_id, appeal_id, need_to_remove, updated_at FROM slots WHERE employee_id = $1 AND need_to_remove = TRUE`, employeeID)
	if err != nil {
		return model.Slot{}, fmt.Errorf("slotRepo.GetNeedToRemoveSlot: %w", err)
	}

	return slot.toDomain(), nil
}
