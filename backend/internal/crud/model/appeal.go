package model

import "time"

type Appeal struct {
	ID         int       `json:"id"`
	ClientID   int       `json:"clientId"`
	EmployeeID *int      `json:"employeeId"`
	ThemeID    int       `json:"themeId"`
	SubthemeID *int      `json:"subthemeId"`
	Text       string    `json:"text"`
	Status     string    `json:"status"`
	TeamID     *int      `json:"teamId"`
	CreatedAt  time.Time `json:"createdAt"`
}
