package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNextMap(t *testing.T) {
	tests := []struct {
		name  string
		edges []Edge
		wantA string
		wantB string
	}{
		{
			name: "two edges",
			edges: []Edge{
				{Source: "a", Target: "b"},
				{Source: "b", Target: "c"},
			},
			wantA: "b",
			wantB: "c",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := buildNextMap(tc.edges)
			assert.Equal(t, tc.wantA, m["a"])
			assert.Equal(t, tc.wantB, m["b"])
		})
	}
}

func TestFindStartNode(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []Node
		nextMap map[string]string
		want    string
	}{
		{
			name:    "has start node",
			nodes:   []Node{{ID: "a"}, {ID: "b"}, {ID: "c"}},
			nextMap: map[string]string{"a": "b", "b": "c"},
			want:    "a",
		},
		{
			name:    "empty input",
			nodes:   nil,
			nextMap: map[string]string{},
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, findStartNode(tc.nodes, tc.nextMap))
		})
	}
}

func strPtr(s string) *string { return &s }

func TestBuildChainPredicateAndAction(t *testing.T) {
	attr := ThemeId
	cmp := Eq
	actionType := AssignTeamAction
	data := ActionData{Values: []string{"5"}}

	nodes := []Node{
		{ID: "pred", Type: PredicateNode, Data: ptrAny(Predicate{Attribute: &attr, Comparison: &cmp, Values: []string{"1"}})},
		{ID: "act", Type: ActionNode, Data: ptrAny(Action{ActionType: &actionType, Data: &data})},
	}
	edges := []Edge{{Source: "pred", Target: "act"}}
	nextMap := buildNextMap(edges)

	svc := WorkflowService{}
	chain := svc.buildChain(nodes, nextMap, "pred")
	assert.NotNil(t, chain)

	result := chain.Do(map[string]interface{}{"themeId": 1})
	assert.Equal(t, BlockResult{}, result)
	next := chain.GetNext()
	assert.NotNil(t, next)
}

func ptrAny(v any) *interface{} {
	var i interface{} = v
	return &i
}

func TestBuildChainConditionNode(t *testing.T) {
	attr := Text
	cmp := Contains
	nodes := []Node{
		{ID: "cond", Type: ConditionNode, Data: ptrAny(ConditionGroup{
			Operator:   ConditionGroupOperatorAnd,
			Conditions: []Predicate{{Attribute: &attr, Comparison: &cmp, Values: []string{"help"}}},
		})},
	}
	svc := WorkflowService{}
	chain := svc.buildChain(nodes, map[string]string{}, "cond")
	assert.NotNil(t, chain)
	chain.Do(map[string]interface{}{"text": "need help"})
}

func TestBuildChainSkipsInvalidNodes(t *testing.T) {
	nodes := []Node{
		{ID: "bad", Type: ActionNode, Data: ptrAny(struct{}{})},
		{ID: "pred", Type: PredicateNode, Data: nil},
	}
	svc := WorkflowService{}
	chain := svc.buildChain(nodes, map[string]string{}, "bad")
	assert.Nil(t, chain)
}
