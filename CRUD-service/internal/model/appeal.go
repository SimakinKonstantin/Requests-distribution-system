package model

// Appeal represents a client's support request.
// EmployeeID is nil when no employee has been assigned yet.
type Appeal struct {
	ID         int    `json:"id"`
	ClientID   int    `json:"clientId"`
	EmployeeID *int   `json:"employeeId"` // nil = not yet assigned
	ThemeID    int    `json:"themeId"`
	SubthemeID int    `json:"subthemeId"`
	Text       string `json:"text"`
	Status     string `json:"status"` // "active" | "closed"
}
