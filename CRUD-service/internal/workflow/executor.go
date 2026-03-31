package workflow

import (
	"workflow-service/gen"
)

type Executor struct {
	status        gen.WorkflowStatus
	chain         actionBlock
	backEndClient *genclient.ClientWithResponses
}

func newWorkflowExecutor(chain actionBlock, backEndClient *genclient.ClientWithResponses, status gen.WorkflowStatus) *Executor {
	return &Executor{
		chain:         chain,
		backEndClient: backEndClient,
		status:        status,
	}
}

func (we *Executor) ExecuteWorkflow(data map[string]interface{}) actionBlockResult {
	current := we.chain
	var result actionBlockResult

	if we.status != gen.Live {
		return result
	}

	for current != nil {
		result = current.Do(data)
		current = current.GetNext()
	}

	return result
}
