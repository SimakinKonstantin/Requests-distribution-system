package workflow

import (
	"testing"
	"time"
	"workflow-service/gen"
)

// helper to format a time in RFC3339 with UTC date portion; date doesn't matter because function checks only time
func timeHelper(s string) string {
	// parse using today's date to build a full RFC3339 timestamp in UTC
	// input s expected like "15:04"
	now := time.Now().UTC()
	parsed, _ := time.Parse("15:04", s)
	combined := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, time.UTC)
	return combined.Format(time.RFC3339)
}

func TestTimeInInterval_Overnight(t *testing.T) {
	// interval 19:00 - 08:00 (overnight)
	start := timeHelper("19:00")
	end := timeHelper("08:00")

	in, err := timeInInterval(timeHelper("20:00"), []string{start, end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !in {
		t.Fatalf("expected 20:00 to be inside 19:00-08:00")
	}

	in2, err := timeInInterval(timeHelper("10:00"), []string{start, end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in2 {
		t.Fatalf("expected 10:00 to be outside 19:00-08:00")
	}
}

func TestTimeInInterval_Normal(t *testing.T) {
	// interval 09:00 - 17:00 (normal)
	start := timeHelper("09:00")
	end := timeHelper("17:00")

	in, err := timeInInterval(timeHelper("10:00"), []string{start, end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !in {
		t.Fatalf("expected 10:00 to be inside 09:00-17:00")
	}

	in2, err := timeInInterval(timeHelper("08:00"), []string{start, end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in2 {
		t.Fatalf("expected 08:00 to be outside 09:00-17:00")
	}
}

func Test_equal(t *testing.T) {
	type args struct {
		valueToCompare interface{}
		values         []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string equal",
			args: args{valueToCompare: "hello", values: []string{"hello"}},
			want: true,
		},
		{
			name: "string not equal",
			args: args{valueToCompare: "hello", values: []string{"world"}},
			want: false,
		},
		{
			name: "int equal",
			args: args{valueToCompare: 123, values: []string{"123"}},
			want: true,
		},
		{
			name: "int not equal",
			args: args{valueToCompare: 123, values: []string{"456"}},
			want: false,
		},
		{
			name: "unsupported type",
			args: args{valueToCompare: 123.45, values: []string{"123.45"}},
			want: false,
		},
		{
			name: "no values",
			args: args{valueToCompare: "hello", values: []string{}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := equal(tt.args.valueToCompare, tt.args.values); got != tt.want {
				t.Errorf("equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contains(t *testing.T) {
	type args struct {
		valueToCompare interface{}
		values         []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string contains",
			args: args{valueToCompare: "hello world", values: []string{"world"}},
			want: true,
		},
		{
			name: "string not contains",
			args: args{valueToCompare: "hello world", values: []string{"test"}},
			want: false,
		},
		{
			name: "multiple values, one contains",
			args: args{valueToCompare: "hello world", values: []string{"test", "world"}},
			want: true,
		},
		{
			name: "unsupported type",
			args: args{valueToCompare: 123, values: []string{"123"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := contains(tt.args.valueToCompare, tt.args.values); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_interval(t *testing.T) {
	type args struct {
		valueToCompare interface{}
		values         []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "value in interval",
			args: args{valueToCompare: "2025-10-10T01:00:00+03:00", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T23:59:59+03:00"}},
			want: true,
		},
		{
			name: "value in interval (==end)",
			args: args{valueToCompare: "2025-10-10T23:59:59+03:00", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T23:59:59+03:00"}},
			want: true,
		},
		{
			name: "value in interval (==start)",
			args: args{valueToCompare: "2025-10-10T00:00:00+03:00", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T23:59:59+03:00"}},
			want: true,
		},
		{
			name: "value in interval (start==end)",
			args: args{valueToCompare: "2025-10-10T00:00:00+03:00", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T00:00:00+03:00"}},
			want: true,
		},
		{
			name: "value not in interval",
			args: args{valueToCompare: "2025-12-12T00:00:00+03:00", values: []string{
				"2025-10-10T01:00:00+03:00", "2025-10-10T02:00:00+03:00", "2025-11-10T10:00:00+03:00", "2025-11-10T11:00:00+03:00"}},
			want: false,
		},
		{
			name: "incorrect time value",
			args: args{valueToCompare: "qwerty", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T00:01:00+03:00"}},
			want: false,
		},
		{
			name: "incorrect interval (start > end)",
			args: args{valueToCompare: "2025-10-10T00:00:00+03:00", values: []string{"2025-10-10T23:59:59", "2025-10-10T00:00:00+03:00"}},
			want: false,
		},
		{
			name: "incorrect interval",
			args: args{valueToCompare: "2025-10-10T00:00:00+03:00", values: []string{"2025-10-10T00:00:00+03:00", "2025-10-10T02:00:00+03:00", "2025-10-10T03:00:00+03:00"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := timeInInterval(tt.args.valueToCompare, tt.args.values); got != tt.want {
				t.Errorf("timeInInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPredicateBlock_Do(t *testing.T) {
	type fields struct {
		predicate gen.Predicate
	}
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name           string
		fields         fields
		attribute      gen.PredicateAttribute
		comparison     gen.PredicateComparison
		args           args
		wantSkipNext   bool
		predicateAfter gen.Predicate
	}{
		{
			name:       "eq - success",
			attribute:  gen.ClientEmail,
			comparison: gen.Eq,
			fields: fields{
				predicate: gen.Predicate{
					Values: []string{"n.kolchin@smartway.today"},
				},
			},
			args: args{
				data: map[string]interface{}{
					"clientEmail": "n.kolchin@smartway.today",
				},
			},
			wantSkipNext: false,
		},
		{
			name:       "eq - fail",
			attribute:  gen.ClientEmail,
			comparison: gen.Eq,
			fields: fields{
				predicate: gen.Predicate{
					Values: []string{"n.kolchin@smartway.today"},
				},
			},
			args: args{
				data: map[string]interface{}{
					"clientEmail": "another@email.com",
				},
			},
			wantSkipNext: true,
		},
		{
			name:       "attribute not found",
			attribute:  "nonExistent",
			comparison: gen.Eq,
			fields: fields{
				predicate: gen.Predicate{
					Values: []string{"value"},
				},
			},
			args: args{
				data: map[string]interface{}{
					"clientEmail": "n.kolchin@smartway.today",
				},
			},
			wantSkipNext: true,
		},
		{
			name:       "unknown comparison",
			attribute:  gen.ClientEmail,
			comparison: "unknown",
			fields: fields{
				predicate: gen.Predicate{
					Values: []string{"n.kolchin@smartway.today"},
				},
			},
			args: args{
				data: map[string]interface{}{
					"clientEmail": "n.kolchin@smartway.today",
				},
			},
			wantSkipNext: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.fields.predicate.Attribute = &tt.attribute
			tt.fields.predicate.Comparison = &tt.comparison

			p := &predicateBlock{
				predicate: tt.fields.predicate,
			}
			p.Do(tt.args.data)
			if p.skipNext != tt.wantSkipNext {
				t.Errorf("PredicateBlock.Do() skipNext = %v, want %v", p.skipNext, tt.wantSkipNext)
			}
		})
	}
}

func TestIsDayInPeriods(t *testing.T) {
	// choose concrete dates: 2025-10-04 is Saturday (weekend), 2025-10-06 is Monday (weekday)
	weekendDate := time.Date(2025, 10, 4, 12, 0, 0, 0, time.UTC)
	weekdayDate := time.Date(2025, 10, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     time.Time
		periods  []string
		isDayoff bool // value returned from external API (true for dayoff/weekend)
		want     bool
	}{
		{name: "weekend true", date: weekendDate, periods: []string{"weekend"}, isDayoff: true, want: true},
		{name: "weekend false", date: weekendDate, periods: []string{"weekend"}, isDayoff: false, want: false},
		{name: "weekday true", date: weekdayDate, periods: []string{"weekday"}, isDayoff: false, want: true},
		{name: "weekday false (dayoff)", date: weekdayDate, periods: []string{"weekday"}, isDayoff: true, want: false},
		{name: "both present", date: weekendDate, periods: []string{"weekday", "weekend"}, isDayoff: false, want: true},
		{name: "both present 2", date: weekdayDate, periods: []string{"weekend", "weekday"}, isDayoff: true, want: true},
		{name: "case insensitive and spaces", date: weekendDate, periods: []string{"  WeekEnd  "}, isDayoff: true, want: true},
		{name: "empty periods", date: weekdayDate, periods: []string{}, isDayoff: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isDayInPeriods(tt.date, tt.periods, tt.isDayoff)
			if err != nil {
				// empty periods is expected to return an error according to implementation
				if len(tt.periods) == 0 {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("isDayInPeriods() = %v, want %v", got, tt.want)
			}
		})
	}
}
