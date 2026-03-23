package model

// Slot links an Employee to an Appeal (one active appeal per slot).
type Slot struct {
	ID           int  `json:"id"`
	EmployeeID   int  `json:"employeeId"`
	AppealID     *int `json:"appealId"`
	NeedToRemove bool `json:"needToRemove"`
}
