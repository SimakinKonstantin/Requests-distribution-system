package model

import "time"

type PendingAppeal struct {
	AppealID  int       `json:"appealId"`
	TeamID    int       `json:"teamId"`
	Priority  int       `json:"priority"`
	UpdatedAt time.Time `json:"updatedAt"`
}
