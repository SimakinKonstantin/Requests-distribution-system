package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEqual(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		values    []string
		want      bool
		shouldErr bool
	}{
		{name: "exact match", value: "test@mail.ru", values: []string{"test@mail.ru"}, want: true},
		{name: "case insensitive", value: "Test@Mail.RU", values: []string{"test@mail.ru"}, want: true},
		{name: "int match", value: 42, values: []string{"42"}, want: true},
		{name: "no match", value: "other", values: []string{"test@mail.ru"}, want: false},
		{name: "empty values", value: "x", values: nil, shouldErr: true},
		{name: "unsupported type", value: true, values: []string{"true"}, shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := equal(tc.value, tc.values)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		values    []string
		want      bool
		shouldErr bool
	}{
		{name: "contains", value: "Hello World", values: []string{"world"}, want: true},
		{name: "does not contain", value: "Hello", values: []string{"xyz"}, want: false},
		{name: "unsupported type", value: 123, values: []string{"1"}, shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := contains(tc.value, tc.values)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestTimeInInterval(t *testing.T) {
	start := "2024-01-01T09:00:00Z"
	end := "2024-01-01T18:00:00Z"

	overnightStart := "2024-01-01T22:00:00Z"
	overnightEnd := "2024-01-01T06:00:00Z"
	tests := []struct {
		name      string
		value     interface{}
		intervals []string
		want      bool
		shouldErr bool
	}{
		{name: "inside interval", value: "2024-05-15T12:00:00Z", intervals: []string{start, end}, want: true},
		{name: "outside interval", value: "2024-05-15T08:00:00Z", intervals: []string{start, end}, want: false},
		{name: "odd intervals", value: "2024-05-15T12:00:00Z", intervals: []string{start}, shouldErr: true},
		{name: "invalid datetime", value: "not-a-time", intervals: []string{start, end}, shouldErr: true},
		{name: "overnight interval", value: "2024-05-15T23:00:00Z", intervals: []string{overnightStart, overnightEnd}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := timeInInterval(tc.value, tc.intervals)
			if tc.shouldErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestGetOnlyTime(t *testing.T) {
	tests := []struct {
		name   string
		input  time.Time
		hour   int
		minute int
	}{
		{
			name:   "extract hour and minute",
			input:  time.Date(2024, 5, 15, 14, 30, 45, 0, time.UTC),
			hour:   14,
			minute: 30,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getOnlyTime(tc.input)
			assert.Equal(t, tc.hour, got.Hour())
			assert.Equal(t, tc.minute, got.Minute())
		})
	}
}

func TestPredicateBlock(t *testing.T) {
	attr := ClientEmail
	cmp := Eq
	block := newPredicateBlock(Predicate{
		Attribute:  &attr,
		Comparison: &cmp,
		Values:     []string{"vip@test.ru"},
	})

	result := block.Do(map[string]interface{}{"clientEmail": "vip@test.ru"})
	assert.False(t, block.skipNext)
	assert.Equal(t, BlockResult{}, result)

	block = newPredicateBlock(Predicate{
		Attribute:  &attr,
		Comparison: &cmp,
		Values:     []string{"vip@test.ru"},
	})
	block.Do(map[string]interface{}{"clientEmail": "other@test.ru"})
	assert.True(t, block.skipNext)

	allCmp := All
	block = newPredicateBlock(Predicate{Attribute: &attr, Comparison: &allCmp})
	block.Do(map[string]interface{}{"clientEmail": "any"})
	assert.False(t, block.skipNext)

	block = newPredicateBlock(Predicate{})
	block.Do(map[string]interface{}{})
	assert.False(t, block.skipNext)

	next := newPredicateBlock(Predicate{})
	block.SetNext(next)
	assert.Equal(t, next, block.GetNext())
	block.skipNext = true
	assert.Nil(t, block.GetNext())
	assert.True(t, block.End())
}
