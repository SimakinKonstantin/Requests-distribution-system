package balancer

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
	AppealID int `json:"appealId"`
	TeamID   int `json:"teamId"`
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

// AppealRow mirrors the columns read from the appeals table by the balancer.
type AppealRow struct {
	ID                     int
	TeamID                 int
	IsUrgent               bool
	IsImportant            bool
	CreatedAt              time.Time
	PendingClientMessageAt *time.Time
	ManagerID              *int // maps to appeals.employee_id
	Status                 string
}

// ManagerRow mirrors the columns read from the employees table by the balancer.
type ManagerRow struct {
	ID            int
	IsAvailable   bool
	TeamIDs       []int
	ActiveAppeals int
	LastAssignAt  *time.Time
}

// SlotRow mirrors the columns read from the slots table by the balancer.
type SlotRow struct {
	ID        int
	ManagerID int // maps to slots.employee_id
	AppealID  *int
	UpdatedAt time.Time
}
