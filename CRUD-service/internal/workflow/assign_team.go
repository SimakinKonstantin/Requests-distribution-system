package workflow

import (
	"context"
	"crud-service/internal/crud/service"
	"fmt"
	"log/slog"
	"strconv"
)

type actionBlockAssignTeam struct {
	ctx         context.Context
	next        actionBlock
	teamId      int
	teamService service.TeamService
}

func newActionBlockAssignTeam(values []string) *actionBlockAssignTeam {
	if len(values) == 0 {
		slog.Error("provide empty values")
		return &actionBlockAssignTeam{}
	}

	teamID, err := strconv.Atoi(values[0])
	if err != nil {
		slog.Error(fmt.Sprintf("error to convert team id: %w", err))
		return &actionBlockAssignTeam{}
	}

	return &actionBlockAssignTeam{
		ctx:    context.Background(),
		teamId: teamID,
	}
}

func (a *actionBlockAssignTeam) Do(data map[string]interface{}) BlockResult {
	val, ok := data["appealId"]
	if !ok {
		return BlockResult{}
	}

	appealId, ok := val.(int)
	if !ok {
		return BlockResult{}
	}

	err := a.teamService.AssignTeam(int(appealId), int(a.teamId))
	if err != nil {
		slog.Error(fmt.Sprintf("error to send request: %w", err))
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
