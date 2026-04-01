package workflow

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

func (we *Executor) ExecuteWorkflow(data map[string]interface{}) BlockResult {
	current := we.chain
	var result BlockResult

	if we.status != StatusLive {
		return result
	}

	for current != nil {
		result = current.Do(data)
		current = current.GetNext()
	}

	return result
}
