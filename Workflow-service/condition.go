package workflow

import (
	"fmt"
	"log/slog"
	"workflow-service/gen"
)

var (
	_attributePlurals = map[gen.PredicateAttribute]string{
		gen.ClientEmail: "clientEmails",
	}
)

type conditionBlock struct {
	conditionGroup gen.ConditionGroup
	skipNext       bool
	next           actionBlock
}

func newConditionBlock(conditionGroup gen.ConditionGroup) *conditionBlock {
	return &conditionBlock{conditionGroup: conditionGroup}
}

func (c *conditionBlock) Do(data map[string]interface{}) actionBlockResult {
	conditionGroupResult, err := initBoolResult(c.conditionGroup.Operator)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка обработки условия: %s", err))
		c.skipNext = true
		return actionBlockResult{}
	}

	// Обход всех подгрупп условий.
	for _, condition := range c.conditionGroup.Conditions {
		conditionResult, err := initBoolResult(gen.ConditionGroupOperator(condition.Operator))
		if err != nil {
			slog.Error("Ошибка обработки условия из группы: %s", err)
			c.skipNext = true
			return actionBlockResult{}
		}

		// Обход всех условий внутри condition.
		for _, predicate := range condition.Predicates {
			predicateResult := c.checkPredicate(predicate, data)

			conditionResult, err = changeBoolResult(conditionResult, predicateResult, gen.ConditionGroupOperator(condition.Operator))
			if err != nil {
				slog.Error("Ошибка обработки предиката: %s", err)
				c.skipNext = true
				return actionBlockResult{}
			}
		}

		conditionGroupResult, err = changeBoolResult(conditionGroupResult, conditionResult, c.conditionGroup.Operator)
		if err != nil {
			slog.Error("Ошибка обработки условия из группы: %s", err)
			c.skipNext = true
			return actionBlockResult{}
		}
	}

	c.skipNext = !conditionGroupResult

	return actionBlockResult{}
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
func initBoolResult(operator gen.ConditionGroupOperator) (bool, error) {
	if operator == gen.ConditionGroupOperatorAnd {
		return true, nil
	}

	if operator == gen.ConditionGroupOperatorOr {
		return false, nil
	}

	return false, fmt.Errorf("неизвестный логический оператор: %s", string(operator))
}

func changeBoolResult(curValue, value bool, operator gen.ConditionGroupOperator) (bool, error) {
	if operator == gen.ConditionGroupOperatorAnd {
		return curValue && value, nil
	}

	if operator == gen.ConditionGroupOperatorOr {
		return curValue || value, nil
	}

	return false, fmt.Errorf("неизвестный логический оператор: %s", string(operator))
}

func (c *conditionBlock) checkPredicate(predicate gen.Predicate, data map[string]interface{}) bool {
	if predicate.Attribute == nil || predicate.Comparison == nil {
		slog.Error("Predicate has nil field")
		return false
	}

	val, ok := data[string(*predicate.Attribute)]
	if !ok {
		val, ok = getAttributePlural(*predicate.Attribute, data)
		if !ok {
			slog.Error("Attribute '%s' not found in data", *predicate.Attribute)
			return false
		}
	}

	if *predicate.Attribute == gen.UserAgent {
		valueStr, ok := val.(string)
		if !ok {
			slog.Error("Attribute is UserAgent, but has no value")
			return false
		}

		val = ParseUserAgent(valueStr)
	}

	var (
		predicateResult bool
		err             error
	)

	switch string(*predicate.Comparison) {
	// case string(gen.PredicateComparisonEq):
	// 	predicateResult, err = anyEqual(val, predicate.Values)
	// case string(gen.PredicateComparisonContains):
	// 	predicateResult, err = contains(val, predicate.Values)
	// case string(gen.PredicateComparisonInInterval):
	// 	predicateResult, err = timeInInterval(val, predicate.Values)
	// case string(gen.PredicateComparisonNotEq):
	// 	predicateResult, err = anyEqual(val, predicate.Values)
	// 	predicateResult = !predicateResult
	// case string(gen.PredicateComparisonNotContains):
	// 	predicateResult, err = contains(val, predicate.Values)
	// 	predicateResult = !predicateResult
	// case string(gen.PredicateComparisonNotInInterval):
	// 	predicateResult, err = timeInInterval(val, predicate.Values)
	// 	predicateResult = !predicateResult
	default:
		err = fmt.Errorf("unknown comparison operator: %s", string(*predicate.Comparison))
		predicateResult = false
	}

	if err != nil {
		// swlog.Global().Error(err.Error())
		return false
	}

	return predicateResult
}

const (
	AndroidUA = "android"
	IosUA     = "ios"
	EmailUA   = "email"
	WebUA     = "web"
)

// Перевод userAgent в формат, который приходит с фронта.
// func ParseUserAgent(userAgent string) string {
// 	agent := useragent.Parse(userAgent)

// 	// todo Убрать ktor-client, когда будет отправляться валидный userAgent с андроида.
// 	if userAgent == "ktor-client" {
// 		return AndroidUA
// 	}

// 	// todo Пока для проверки на ios, т.к. либа некорректно парсит для ios
// 	if strings.HasPrefix(userAgent, "SmartwayBeta") || strings.HasPrefix(userAgent, "Smartway") {
// 		return IosUA
// 	}

// 	// Если agent отправляется через mailService, то будет установлен некорректный userAgent.
// 	if agent.IsUnknown() {
// 		return EmailUA
// 	}

// 	if agent.OS == useragent.IOS {
// 		return IosUA
// 	}

// 	if agent.OS == useragent.Android {
// 		return AndroidUA
// 	}

// 	return WebUA
// }

func getAttributePlural(attribute gen.PredicateAttribute, data map[string]any) (any, bool) {
	var (
		plural string
		exists bool
	)
	if plural, exists = _attributePlurals[attribute]; !exists {
		return nil, false
	}

	if alt, found := data[plural]; found {
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
