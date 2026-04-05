package workflow

import (
	"fmt"
	"log/slog"
)

type conditionBlock struct {
	conditionGroup ConditionGroup
	skipNext       bool
	next           actionBlock
}

func newConditionBlock(conditionGroup ConditionGroup) *conditionBlock {
	return &conditionBlock{conditionGroup: conditionGroup}
}

func (c *conditionBlock) Do(payload map[string]interface{}) BlockResult {
	conditionGroupResult, err := initBoolResult(c.conditionGroup.Operator)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка обработки условия: %s", err))
		c.skipNext = true
		return BlockResult{}
	}

	for _, condition := range c.conditionGroup.Conditions {
		slog.Warn(fmt.Sprintf("CONDITION group result before: %+v", conditionGroupResult))

		predicateResult := c.checkPredicate(condition, payload)

		slog.Warn(fmt.Sprintf("PREDICATE RESULT: %v", predicateResult))

		conditionGroupResult, err = changeBoolResult(conditionGroupResult, predicateResult, ConditionGroupOperator(c.conditionGroup.Operator))
		if err != nil {
			slog.Error("Ошибка обработки условия из группы: %s", err.Error())
			c.skipNext = true
			return BlockResult{}
		}

		slog.Warn(fmt.Sprintf("CONDITION GROUP RESULT: %v", conditionGroupResult))
	}

	c.skipNext = !conditionGroupResult

	return BlockResult{}
}

func (c *conditionBlock) GetNext() actionBlock {
	if c.skipNext {
		return nil
	}
	return c.next
}

func (c *conditionBlock) SetNext(next actionBlock) {
	c.next = next
}

func (c *conditionBlock) End() bool {
	return c.skipNext || c.next == nil
}

// Инициализирует значение-результат для логического условия.
func initBoolResult(operator ConditionGroupOperator) (bool, error) {
	if operator == ConditionGroupOperatorAnd {
		return true, nil
	}

	if operator == ConditionGroupOperatorOr {
		return false, nil
	}

	return false, fmt.Errorf("неизвестный логический оператор: %s", string(operator))
}

func changeBoolResult(curValue, value bool, operator ConditionGroupOperator) (bool, error) {
	if operator == ConditionGroupOperatorAnd {
		return curValue && value, nil
	}

	if operator == ConditionGroupOperatorOr {
		return curValue || value, nil
	}

	return false, fmt.Errorf("неизвестный логический оператор: %s", string(operator))
}

func (c *conditionBlock) checkPredicate(predicate Predicate, payload map[string]interface{}) bool {

	slog.Warn(fmt.Sprintf("CHECKING PREDICATE WITH VALUES: %+v", predicate.Values))

	if predicate.Attribute == nil || predicate.Comparison == nil {
		slog.Error("Predicate has nil field")
		return false
	}

	val, ok := payload[string(*predicate.Attribute)]
	if !ok {
		val, ok = getAttributePlural(*predicate.Attribute, payload)
		if !ok {
			slog.Error("Attribute '%s' not found in data", *predicate.Attribute)
			return false
		}
	}

	var (
		predicateResult bool
		err             error
	)

	switch string(*predicate.Comparison) {
	case string(Eq):
		predicateResult, err = anyEqual(val, predicate.Values)
	case string(Contains):
		predicateResult, err = contains(val, predicate.Values)
	case string(InInterval):
		predicateResult, err = timeInInterval(val, predicate.Values)
	case string(NotEq):
		predicateResult, err = anyEqual(val, predicate.Values)
		predicateResult = !predicateResult
	case string(NotContains):
		predicateResult, err = contains(val, predicate.Values)
		predicateResult = !predicateResult
	case string(NotInInterval):
		predicateResult, err = timeInInterval(val, predicate.Values)
		predicateResult = !predicateResult
	default:
		err = fmt.Errorf("unknown comparison operator: %s", string(*predicate.Comparison))
		predicateResult = false
	}

	if err != nil {
		slog.Error(fmt.Sprintf("Error checking predicate: %s", err.Error()))
		return false
	}

	return predicateResult
}

func getAttributePlural(attribute PredicateAttribute, payload map[string]any) (any, bool) {

	if alt, found := payload[string(attribute)]; found {
		return alt, true
	}

	return nil, false
}

func anyEqual(valuesToCompare any, values []string) (bool, error) {
	switch vv := valuesToCompare.(type) {
	case []string:
		return sliceEqual(vv, values)
	case []int:
		return sliceEqual(vv, values)
	case []any:
		return sliceEqual(vv, values)
	default:
		return equal(valuesToCompare, values)
	}
}

func sliceEqual[T any](items []T, values []string) (bool, error) {
	for _, item := range items {
		ok, err := equal(item, values)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
