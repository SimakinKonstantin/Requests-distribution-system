package model

import "time"

type Employee struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Surname      string     `json:"surname"`
	Limit        int        `json:"limit"`
	Email        string     `json:"email"`
	Status       string     `json:"status"`
	TeamIDs      []int      `json:"teamIds"`
	LastAssignAt *time.Time `json:"lastAssignAt"`
}
