package model

import "time"

type PendingAppeal struct {
	AppealID  int       `json:"appealId"`
	Priority  int       `json:"priority"`
	UpdatedAt time.Time `json:"updatedAt"`
}
