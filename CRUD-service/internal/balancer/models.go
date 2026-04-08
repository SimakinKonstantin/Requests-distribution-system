package main

import "time"

type EventName string

const (
	EventAppealNeedsDistribution EventName = "APPEAL_NEEDS_DISTRIBUTION"
	EventAppealClosed            EventName = "APPEAL_CLOSED"
)

type RabbitEvent struct {
	Name    EventName       `json:"name"`
	Payload RabbitEventBody `json:"payload"`
}

type RabbitEventBody struct {
	AppealID int    `json:"appealId"`
	TeamID   string `json:"teamId"`
}

type JobBatchType string

const (
	BatchAppealStatusChanges       JobBatchType = "appealStatusChanges"
	BatchManagerAvailabilityChange JobBatchType = "managerAvailabilityChanges"
	BatchDistributionRequests      JobBatchType = "distributionRequests"
)

type ProcessedEvent struct {
	RabbitEvent
	ProcessedAt time.Time `json:"processedAt"`
}

type AppealRow struct {
	ID                         int
	TeamID                     string
	IsUrgent                   bool
	IsImportant                bool
	CreatedAt                  time.Time
	PendingClientMessageAt     *time.Time
	ManagerID                  *string
	Status                     string
}

type ManagerRow struct {
	ID              string
	IsAvailable     bool
	TeamIDs         []string
	ActiveAppeals   int
	LastAssignAt    *time.Time
}

type SlotRow struct {
	ID        string
	ManagerID string
	AppealID  *int
	UpdatedAt time.Time
}

