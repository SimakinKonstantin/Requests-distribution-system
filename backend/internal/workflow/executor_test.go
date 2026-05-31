package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutorPausedWorkflow(t *testing.T) {
	attr := ThemeId
	cmp := Eq
	tests := []struct {
		name   string
		status Status
	}{
		{name: "paused", status: StatusPaused},
		{name: "live", status: StatusLive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block := newPredicateBlock(Predicate{Attribute: &attr, Comparison: &cmp, Values: []string{"1"}})
			executor := newWorkflowExecutor(block, tc.status)
			result := executor.ExecuteWorkflow(map[string]interface{}{"themeId": 1})
			assert.Equal(t, BlockResult{}, result)
		})
	}
}
