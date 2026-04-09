package balancer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

type Assigner struct {
	db *DB
}

func NewAssigner(db *DB) *Assigner { return &Assigner{db: db} }

func (a *Assigner) HandleAssignTask(ctx context.Context, t *asynq.Task) error {
	var p AssignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	err := a.db.AssignAppealTx(ctx, p.AppealID, p.ManagerID, p.SlotID)
	if err == ErrAppealAlreadyAssigned {
		return nil // idempotent
	}
	if err == ErrNoFreeSlot {
		log.Printf("assigner: no free slot for manager=%d slot=%d appeal=%d", p.ManagerID, p.SlotID, p.AppealID)
		return err
	}
	if err != nil {
		return err
	}

	log.Printf("assigner: assigned appeal=%d manager=%d slot=%d team=%d", p.AppealID, p.ManagerID, p.SlotID, p.TeamID)
	return nil
}
