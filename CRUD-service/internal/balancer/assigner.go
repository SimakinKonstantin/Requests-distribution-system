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
		// idempotent
		return nil
	}
	if err == ErrNoFreeSlot {
		// allow retry: maybe another tick will find different slot
		log.Printf("assigner: no free slot for manager=%s slot=%s appeal=%d", p.ManagerID, p.SlotID, p.AppealID)
		return err
	}
	if err != nil {
		return err
	}

	log.Printf("assigner: assigned appeal=%d manager=%s slot=%s team=%s", p.AppealID, p.ManagerID, p.SlotID, p.TeamID)
	return nil
}

