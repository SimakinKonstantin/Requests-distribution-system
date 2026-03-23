package workflow

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"workflow-service/gen"
)

type predicateBlock struct {
	predicate gen.Predicate
	next      actionBlock
	skipNext  bool
}

const (
	weekday = "weekday"
	weekend = "weekend"
)

func newPredicateBlock(predicate gen.Predicate) *predicateBlock {
	return &predicateBlock{predicate: predicate}
}

func (p *predicateBlock) Do(data map[string]interface{}) actionBlockResult {
	if p.predicate.Attribute == nil || p.predicate.Comparison == nil {
		return actionBlockResult{}
	}

	val, ok := data[string(*p.predicate.Attribute)]

	if !ok {
		fmt.Printf("Attribute '%s' not found in data", string(*p.predicate.Attribute))
		p.skipNext = true
		return actionBlockResult{}
	}

	var (
		predicateResult bool
		err             error
	)

	switch string(*p.predicate.Comparison) {
	case string(gen.PredicateComparisonEq):
		predicateResult, err = equal(val, p.predicate.Values)
	case string(gen.PredicateComparisonNotEq):
		predicateResult, err = equal(val, p.predicate.Values)
		predicateResult = !predicateResult
	case string(gen.PredicateComparisonAll):
		p.skipNext = false
		return actionBlockResult{}
	default:
		err = fmt.Errorf("unknown comparison operator: %s", string(*p.predicate.Comparison))
		predicateResult = false
	}

	if err != nil {
		slog.Error(err.Error())
		p.skipNext = true
		return actionBlockResult{}
	}

	p.skipNext = !predicateResult
	return actionBlockResult{}
}

func (p *predicateBlock) GetNext() actionBlock {
	if p.skipNext {
		return nil
	}
	return p.next
}

func (p *predicateBlock) SetNext(next actionBlock) {
	p.next = next
}

func (p *predicateBlock) End() bool {
	return p.skipNext || p.next == nil
}

func equal(valueToCompare interface{}, values []string) (bool, error) {
	if len(values) == 0 {
		return false, errors.New("no values to compare")
	}

	strValue := ""
	switch v := valueToCompare.(type) {
	case string:
		strValue = v
	case int:
		strValue = strconv.Itoa(v)
	default:
		return false, errors.New("unsupported attribute value type to compare")
	}

	for _, value := range values {
		if strings.EqualFold(strValue, value) {
			return true, nil
		}
	}
	return false, nil
}

func contains(valueToCompare interface{}, values []string) (bool, error) {
	value, ok := valueToCompare.(string)
	if !ok {
		return false, errors.New("unsupported attribute value type to check containing")
	}

	for _, v := range values {
		if strings.Contains(strings.ToLower(value), strings.ToLower(v)) {
			return true, nil
		}
	}
	return false, nil
}

// Проверка, что время входит в какой-то интервал
func timeInInterval(timeVal interface{}, intervals []string) (bool, error) {
	if len(intervals)%2 != 0 {
		return false, fmt.Errorf("intervals must have an even number of intervals, len: %d", len(intervals))
	}

	parsedDateTime, err := parseDateTime(timeVal)
	if err != nil {
		return false, fmt.Errorf("cant parse time value: %s", err)
	}

	// Время проверяется, как cron (т.е. дата не проверяется).
	parsedTime := getOnlyTime(parsedDateTime)

	for i := 0; i < len(intervals); i += 2 {
		startDateTime, err := time.Parse(time.RFC3339, intervals[i])
		if err != nil {
			return false, fmt.Errorf("cant parse time value from boundary: %s", err)
		}
		startTime := getOnlyTime(startDateTime)

		endDateTime, err := time.Parse(time.RFC3339, intervals[i+1])
		if err != nil {
			return false, fmt.Errorf("cant parse time value from boundary: %s", err)
		}
		endTime := getOnlyTime(endDateTime)

		if startTime.Before(endTime) && !parsedTime.Before(startTime) && !parsedTime.After(endTime) {
			return true, nil
		}

		if startTime.Equal(endTime) && parsedTime.Equal(startTime) {
			return true, nil
		}

		if startTime.After(endTime) {
			if !parsedTime.Before(startTime) || !parsedTime.After(endTime) {
				return true, nil
			}
		}

		slog.Error(fmt.Sprintf("Время %s не входит в интервал: (%s : %s)", parsedDateTime, intervals[i], intervals[i+1]))
	}

	return false, nil
}

func getOnlyTime(dateTime time.Time) time.Time {
	return time.Date(0, 1, 1, dateTime.Hour(), dateTime.Minute(), 0, 0, time.UTC)
}

func parseDateTime(timeVal interface{}) (time.Time, error) {
	timeToCompare, ok := timeVal.(string)
	if !ok {
		return time.Time{}, errors.New("unsupported attribute value type to check containing")
	}

	parsedDateTime, err := time.Parse(time.RFC3339, timeToCompare)
	if err != nil {
		return time.Time{}, fmt.Errorf("cant parse time value: %s", err)
	}

	return parsedDateTime, nil
}

func isDayInPeriods(date time.Time, periods []string, isDayoff bool) (bool, error) {
	if len(periods) == 0 {
		return false, errors.New("no periods to compare")
	}

	var hasWeekday, hasWeekend bool
	for _, p := range periods {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case weekday:
			hasWeekday = true
		case weekend:
			hasWeekend = true
		}
	}

	if hasWeekday && hasWeekend {
		return true, nil
	}

	if hasWeekend {
		if isDayoff {
			return true, nil
		}
		return false, nil
	}

	if hasWeekday {
		if !isDayoff {
			return true, nil
		}
		return false, nil
	}

	return false, nil
}
