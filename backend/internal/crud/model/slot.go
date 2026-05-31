package model

import "time"

type Slot struct {
	ID           int        `json:"id"`
	EmployeeID   int        `json:"employeeId"`
	AppealID     *int       `json:"appealId"`
	NeedToRemove bool       `json:"needToRemove"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}
