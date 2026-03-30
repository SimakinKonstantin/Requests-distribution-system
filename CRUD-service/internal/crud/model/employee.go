package model

// Employee represents a support-team worker.
type Employee struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Limit   int    `json:"limit"`
	Email   string `json:"email"`
	Status  string `json:"status"`
	TeamIDs []int  `json:"teamIds"`
}
