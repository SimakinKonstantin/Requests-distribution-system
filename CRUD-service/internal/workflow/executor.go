package workflow

import (
	"fmt"
	"log/slog"
)

type Executor struct {
	status Status
	chain  actionBlock
}

func newWorkflowExecutor(chain actionBlock, status Status) *Executor {
	return &Executor{
		chain:  chain,
		status: status,
	}
}

func (we *Executor) ExecuteWorkflow(payload map[string]interface{}) BlockResult {
	slog.Info("Starting to execute workflow")
	current := we.chain
	var result BlockResult

	if we.status != StatusLive {
		return result
	}

	for current != nil {
		slog.Warn(fmt.Sprintf("Executing block: %+v", current))
		result = current.Do(payload)

		slog.Warn(fmt.Sprintf("Executed block: %+v", result))
		current = current.GetNext()
	}

	return result
}
