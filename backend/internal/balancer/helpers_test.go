package balancer

import (
	"testing"
	"time"

	"crud-service/internal/crud/model"

	"github.com/stretchr/testify/assert"
)

func TestCountFreeSlots(t *testing.T) {
	tests := []struct {
		name      string
		state     *employeeState
		freeSlots map[int][]model.Slot
		want      int
	}{
		{
			name:  "counts only unused",
			state: &employeeState{emploee: model.Employee{ID: 1}, usedSlots: map[int]struct{}{1: {}}},
			freeSlots: map[int][]model.Slot{
				1: {{ID: 1}, {ID: 2}, {ID: 3}},
			},
			want: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, countFreeSlots(tc.state, tc.freeSlots))
		})
	}
}

func TestOldestFreeSlotTime(t *testing.T) {
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)
	tests := []struct {
		name      string
		state     *employeeState
		freeSlots map[int][]model.Slot
		want      time.Time
	}{
		{
			name:  "returns oldest updated at",
			state: &employeeState{emploee: model.Employee{ID: 1}},
			freeSlots: map[int][]model.Slot{
				1: {{ID: 1, UpdatedAt: &t1}, {ID: 2, UpdatedAt: &t2}},
			},
			want: t1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, oldestFreeSlotTime(tc.state, tc.freeSlots))
		})
	}
}
