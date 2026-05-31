package workflow

import (
	"context"
	"strconv"
)

type actionBlockAssignTeam struct {
	ctx          context.Context
	next         actionBlock
	teamId       int
	teamAssigner TeamAssigner
}

func newActionBlockAssignTeam(values []string) *actionBlockAssignTeam {
	if len(values) == 0 {
		return &actionBlockAssignTeam{}
	}

	teamID, err := strconv.Atoi(values[0])
	if err != nil {
		return &actionBlockAssignTeam{}
	}

	return &actionBlockAssignTeam{
		ctx:    context.Background(),
		teamId: teamID,
	}
}

func (a *actionBlockAssignTeam) Do(payload map[string]interface{}) BlockResult {
	val, ok := payload["appealId"]
	if !ok {
		return BlockResult{}
	}

	appealId, ok := val.(int)
	if !ok {
		return BlockResult{}
	}

	if a.teamAssigner == nil {
		return BlockResult{}
	}

	err := a.teamAssigner.AssignTeam(int(appealId), int(a.teamId))
	if err != nil {
		return BlockResult{}
	}

	return BlockResult{
		TeamID: a.teamId,
	}
}

func (a *actionBlockAssignTeam) SetNext(next actionBlock) {
	a.next = next
}

func (a *actionBlockAssignTeam) GetNext() actionBlock {
	return a.next
}

func (a *actionBlockAssignTeam) End() bool {
	return a.next == nil
}
