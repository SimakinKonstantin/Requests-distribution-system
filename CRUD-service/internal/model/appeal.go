package model

type Appeal struct {
	ID         int    `json:"id"`
	ClientID   int    `json:"clientId"`
	EmployeeID *int   `json:"employeeId"`
	ThemeID    int    `json:"themeId"`
	SubthemeID *int   `json:"subthemeId"`
	Text       string `json:"text"`
	Status     string `json:"status"` // "active" | "closed"
	TeamID     *int   `json:"teamId"`
}
