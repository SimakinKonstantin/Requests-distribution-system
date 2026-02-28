package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

// slotDB is the database-level representation of Slot.
type slotDB struct {
	ID         int `db:"id"`
	EmployeeID int `db:"employee_id"`
	AppealID   int `db:"appeal_id"`
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
