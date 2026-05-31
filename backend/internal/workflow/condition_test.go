package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitBoolResult(t *testing.T) {
	tests := []struct {
		name      string
		operator  ConditionGroupOperator
		want      bool
		shouldErr bool
	}{
		{name: "and", operator: ConditionGroupOperatorAnd, want: true},
		{name: "or", operator: ConditionGroupOperatorOr, want: false},
		{name: "invalid", operator: "xor", shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := initBoolResult(tc.operator)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestChangeBoolResult(t *testing.T) {
	tests := []struct {
		name      string
		current   bool
		value     bool
		operator  ConditionGroupOperator
		want      bool
		shouldErr bool
	}{
		{name: "and", current: true, value: false, operator: ConditionGroupOperatorAnd, want: false},
		{name: "or", current: false, value: true, operator: ConditionGroupOperatorOr, want: true},
		{name: "invalid", current: true, value: true, operator: "invalid", shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := changeBoolResult(tc.current, tc.value, tc.operator)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConditionBlockAnd(t *testing.T) {
	themeAttr := ThemeId
	textAttr := Text
	eq := Eq
	contains := Contains

	group := ConditionGroup{
		Operator: ConditionGroupOperatorAnd,
		Conditions: []Predicate{
			{Attribute: &themeAttr, Comparison: &eq, Values: []string{"1"}},
			{Attribute: &textAttr, Comparison: &contains, Values: []string{"urgent"}},
		},
	}
	tests := []struct {
		name    string
		payload map[string]interface{}
		wantSkip bool
	}{
		{name: "all conditions true", payload: map[string]interface{}{"themeId": 1, "text": "urgent request"}, wantSkip: false},
		{name: "one condition false", payload: map[string]interface{}{"themeId": 1, "text": "normal request"}, wantSkip: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block := newConditionBlock(group)
			block.Do(tc.payload)
			assert.Equal(t, tc.wantSkip, block.skipNext)
		})
	}
}

func TestConditionBlockOr(t *testing.T) {
	themeAttr := ThemeId
	eq := Eq
	group := ConditionGroup{
		Operator: ConditionGroupOperatorOr,
		Conditions: []Predicate{
			{Attribute: &themeAttr, Comparison: &eq, Values: []string{"1"}},
			{Attribute: &themeAttr, Comparison: &eq, Values: []string{"2"}},
		},
	}
	block := newConditionBlock(group)
	block.Do(map[string]interface{}{"themeId": 2})
	assert.False(t, block.skipNext)
}

func TestConditionBlockNilPredicate(t *testing.T) {
	attr := ThemeId
	block := newConditionBlock(ConditionGroup{
		Operator:   ConditionGroupOperatorAnd,
		Conditions: []Predicate{{Attribute: &attr, Comparison: nil}},
	})
	block.Do(map[string]interface{}{"themeId": 1})
	assert.True(t, block.skipNext)
}

func TestAnyEqualSlice(t *testing.T) {
	tests := []struct {
		name  string
		value any
		check []string
		want  bool
	}{
		{name: "slice string contains", value: []string{"a", "b"}, check: []string{"b"}, want: true},
		{name: "slice int no match", value: []int{1, 2}, check: []string{"3"}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := anyEqual(tc.value, tc.check)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestConditionBlockChain(t *testing.T) {
	block := newConditionBlock(ConditionGroup{Operator: ConditionGroupOperatorAnd})
	next := newPredicateBlock(Predicate{})
	block.SetNext(next)
	assert.Equal(t, next, block.GetNext())
	block.skipNext = true
	assert.Nil(t, block.GetNext())
	assert.True(t, block.End())
}
