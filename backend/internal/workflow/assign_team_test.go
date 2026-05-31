package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTeamAssigner struct {
	calls []struct{ appealID, teamID int }
	err   error
}

func (m *mockTeamAssigner) AssignTeam(appealID, teamID int) error {
	m.calls = append(m.calls, struct{ appealID, teamID int }{appealID, teamID})
	return m.err
}

func TestActionBlockAssignTeam(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   int
	}{
		{name: "valid team id", values: []string{"7"}, want: 7},
		{name: "invalid team id", values: []string{"bad"}, want: 0},
		{name: "empty values", values: nil, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block := newActionBlockAssignTeam(tc.values)
			assert.Equal(t, tc.want, block.teamId)
		})
	}
}

func TestActionBlockAssignTeamDo(t *testing.T) {
	tests := []struct {
		name          string
		payload       map[string]interface{}
		assigner      *mockTeamAssigner
		attachAssigner bool
		want          BlockResult
	}{
		{
			name:           "success",
			payload:        map[string]interface{}{"appealId": 10},
			assigner:       &mockTeamAssigner{},
			attachAssigner: true,
			want:           BlockResult{TeamID: 3},
		},
		{
			name:           "nil assigner",
			payload:        map[string]interface{}{"appealId": 10},
			assigner:       nil,
			attachAssigner: false,
			want:           BlockResult{},
		},
		{
			name:           "empty payload",
			payload:        map[string]interface{}{},
			assigner:       &mockTeamAssigner{},
			attachAssigner: true,
			want:           BlockResult{},
		},
		{
			name:           "assigner error",
			payload:        map[string]interface{}{"appealId": 10},
			assigner:       &mockTeamAssigner{err: errors.New("fail")},
			attachAssigner: true,
			want:           BlockResult{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block := newActionBlockAssignTeam([]string{"3"})
			if tc.attachAssigner {
				block.teamAssigner = tc.assigner
			}
			result := block.Do(tc.payload)
			assert.Equal(t, tc.want, result)
		})
	}

	assigner := &mockTeamAssigner{}
	block := newActionBlockAssignTeam([]string{"3"})
	block.teamAssigner = assigner
	_ = block.Do(map[string]interface{}{"appealId": 10})
	assert.Len(t, assigner.calls, 1)
	assert.Equal(t, 10, assigner.calls[0].appealID)
	assert.Equal(t, 3, assigner.calls[0].teamID)

	next := newPredicateBlock(Predicate{})
	block.SetNext(next)
	assert.Equal(t, next, block.GetNext())
	assert.False(t, block.End())
}
