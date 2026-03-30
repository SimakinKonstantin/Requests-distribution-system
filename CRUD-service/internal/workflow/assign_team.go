package workflow

import (
	"context"
	"log/slog"
)

type actionBlockAssignTeam struct {
	// nodeClient *genclient.ClientWithResponses
	ctx    context.Context
	next   actionBlock
	logger slog.Logger
	teamId string
}

func newActionBlockAssignTeam( /*nodeClient *genclient.ClientWithResponses, */ values []string, logger slog.Logger) *actionBlockAssignTeam {
	if len(values) == 0 {
		logger.Error("provide empty values")
		return &actionBlockAssignTeam{}
	}
	return &actionBlockAssignTeam{
		// nodeClient: nodeClient,
		ctx:    context.Background(),
		teamId: values[0],
	}
}

func (a *actionBlockAssignTeam) Do(data map[string]interface{}) BlockResult {
	// val, ok := data["appealId"]
	// if !ok {
	// 	return actionBlockResult{}
	// }

	// appealId, ok := val.(float64)
	// if !ok {
	// return actionBlockResult{}
	// }

	// resp, err := a.nodeClient.WorkflowHandlerControllerAssignTeamWithResponse(a.ctx, fmt.Sprintf("%v", appealId), a.teamId)
	// if err != nil {
	// 	a.logger.Error(fmt.Sprintf("error to send request: %w", err))
	// 	return actionBlockResult{}
	// }

	// if resp == nil {
	// 	a.logger.Error("nil response")
	// 	return actionBlockResult{}
	// }

	// if resp.JSON200 == nil {
	// 	a.logger.Error(fmt.Sprintf(fmt.Sprintf("nil response model. Body: %s", string(resp.Body))))
	// 	return actionBlockResult{}
	// }

	return BlockResult{
		TeamId: "123",
	}

	{
		// TeamId: resp.JSON200.TeamId,
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
