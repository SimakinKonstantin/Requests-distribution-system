package model

// Appeal represents a client's support request.
type Appeal struct {
	ID         int    `json:"id"`
	ClientID   int    `json:"clientId"`
	EmployeeID int    `json:"employeeId"`
	ThemeID    int    `json:"themeId"`
	SubthemeID int    `json:"subthemeId"`
	Text       string `json:"text"`
}
