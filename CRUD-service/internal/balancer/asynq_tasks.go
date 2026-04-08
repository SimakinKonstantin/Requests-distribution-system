package balancer

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TaskTypeBatchUpdate      = "balancer:batch_update"
	TaskTypeDistributionTick = "balancer:distribution_tick"
	TaskTypeAssign           = "balancer:assign"
)

type BatchUpdatePayload struct {
	Type   JobBatchType     `json:"type"`
	Events []ProcessedEvent `json:"events"`
}

type DistributionTickPayload struct{}

type AssignPayload struct {
	AppealID  int    `json:"appealId"`
	ManagerID string `json:"managerId"`
	SlotID    string `json:"slotId"`
	TeamID    string `json:"teamId"`
	Priority  int    `json:"priority"`
}

func NewBatchUpdateTask(p BatchUpdatePayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskTypeBatchUpdate, b), nil
}

func NewDistributionTickTask() *asynq.Task {
	return asynq.NewTask(TaskTypeDistributionTick, nil)
}

func NewAssignTask(p AssignPayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskTypeAssign, b), nil
}

