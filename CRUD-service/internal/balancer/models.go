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
