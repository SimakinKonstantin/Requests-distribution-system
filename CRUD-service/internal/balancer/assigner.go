package balancer

import (
	"context"
	"crud-service/internal/crud/service"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/hibiken/asynq"
)

type Assigner struct {
	appealService service.AppealService
}

func NewAssigner(appealService service.AppealService) *Assigner {
	return &Assigner{appealService: appealService}
}

func (a *Assigner) HandleAssignTask(ctx context.Context, t *asynq.Task) error {
	var p AssignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	err := a.appealService.Assign(p.AppealID, p.ManagerID, p.SlotID)
	if errors.Is(err, service.ErrAppealAlreadyAssigned) {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		// Slot уже занят или недоступен - повтор задачи не поможет, иначе asynq делает MaxRetry попыток и спамит WARN.
		log.Printf("assigner: no free slot for manager=%d slot=%d appeal=%d (dropping task)", p.ManagerID, p.SlotID, p.AppealID)
		return nil
	}
	if err != nil {
		return err
	}

	log.Printf("assigner: assigned appeal=%d manager=%d slot=%d team=%d", p.AppealID, p.ManagerID, p.SlotID, p.TeamID)
	return nil
}
