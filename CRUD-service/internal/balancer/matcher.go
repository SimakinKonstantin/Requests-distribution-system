package balancer

import (
	"context"
	"crud-service/internal/crud/model"
	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"time"

	"github.com/hibiken/asynq"
)

type employeeState struct {
	emploee            model.Employee
	activeAppealsCount int
	lastAssign         *time.Time
	usedSlots          map[int]struct{}
}

type Matcher struct {
	appealService service.AppealService
	slotService   service.SlotService
	employeeRepo  repository.EmployeeRepository
	asynq         *asynq.Client
	cfg           Config
}

func NewMatcher(asynqClient *asynq.Client, cfg Config, appealService service.AppealService, employeeRepo repository.EmployeeRepository, slotService service.SlotService) *Matcher {
	return &Matcher{appealService: appealService, employeeRepo: employeeRepo, slotService: slotService, asynq: asynqClient, cfg: cfg}
}

func (m *Matcher) RunTicker(ctx context.Context) {
	t := time.NewTicker(m.cfg.MatcherTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_, err := m.asynq.EnqueueContext(ctx, NewDistributionTickTask(),
				asynq.Queue("dist"),
				asynq.MaxRetry(0),
			)
			if err != nil {
				log.Printf("matcher: enqueue tick failed: %v", err)
			}
		}
	}
}

func (m *Matcher) HandleDistributionTick(ctx context.Context, _ *asynq.Task) error {
	log.Printf("Entered distribution tick")

	appeals, err := m.appealService.FetchPendingAppeals(m.cfg.FetchAppealsLimit)

	log.Printf("FetchPendingAppeals: %v", appeals)

	if err != nil {
		log.Printf("matcher: distribution tick: FetchPendingAppeals: %v", err)
		return err
	}

	employees, err := m.employeeRepo.FetchAvailableEmployees(m.cfg.FetchManagersLimit)
	if err != nil {
		log.Printf("matcher: distribution tick: FetchAvailableEmployees: %v", err)
		return err
	}

	log.Printf("Fetched available employees: %v", employees)

	if len(appeals) == 0 || len(employees) == 0 {
		return nil
	}

	employeeIDs := make([]int, 0, len(employees))
	for _, employee := range employees {
		employeeIDs = append(employeeIDs, employee.ID)
	}

	log.Printf("Employee IDs: %v", employeeIDs)

	freeSlots, err := m.slotService.FetchFreeSlotsByEmployees(employeeIDs)
	if err != nil {
		log.Printf("matcher: distribution tick: FetchFreeSlotsByEmployees: %v", err)
		return err
	}

	appealsInfo := make([]model.Appeal, 0, len(appeals))
	for _, appeal := range appeals {
		appealsInfo = append(appealsInfo, appeal.Appeal)
	}
	employeesInfo := make([]model.Employee, 0, len(employees))
	for _, employee := range employees {
		employeesInfo = append(employeesInfo, employee.Employee)
	}

	assignments := m.FindOptimalAssignments(appealsInfo, employeesInfo, freeSlots)
	for _, a := range assignments {
		task, err := NewAssignTask(a)
		if err != nil {
			log.Printf("matcher: make assign task failed: %v", err)
			continue
		}
		_, err = m.asynq.EnqueueContext(ctx, task,
			asynq.Queue("assign"),
			asynq.MaxRetry(5),
		)
		if err != nil {
			log.Printf("matcher: enqueue assign failed: %v", err)
		}
	}
	return nil
}

// FindOptimalAssignments distributes pending appeals to available managers.
func (m *Matcher) FindOptimalAssignments(appeals []model.Appeal, employees []model.Employee, freeSlots map[int][]model.Slot) []AssignPayload {
	byTeam := make(map[int][]*employeeState)
	states := make(map[int]*employeeState, len(employees))
	for _, employee := range employees {

		activeAppealsCount, err := m.employeeRepo.GetEmployeeActiveAppeals(employee.ID)
		if err != nil {
			log.Printf("matcher: get employee active appeals failed: %v", err)
			continue
		}

		st := &employeeState{
			emploee:            employee,
			activeAppealsCount: activeAppealsCount,
			lastAssign:         employee.LastAssignAt,
			usedSlots:          map[int]struct{}{},
		}
		states[employee.ID] = st
		for _, team := range employee.TeamIDs {
			byTeam[team] = append(byTeam[team], st)
		}
	}

	for team, arr := range byTeam {
		sort.SliceStable(arr, func(i, j int) bool {
			a, b := arr[i], arr[j]
			if a.activeAppealsCount != b.activeAppealsCount {
				return a.activeAppealsCount < b.activeAppealsCount
			}
			ai, bi := time.Time{}, time.Time{}
			if a.lastAssign != nil {
				ai = *a.lastAssign
			}
			if b.lastAssign != nil {
				bi = *b.lastAssign
			}
			if !ai.Equal(bi) {
				return ai.Before(bi)
			}
			return len(freeSlots[a.emploee.ID]) > len(freeSlots[b.emploee.ID])
		})
		byTeam[team] = arr
	}

	now := time.Now().UTC()
	out := make([]AssignPayload, 0)

	for _, appeal := range appeals {
		if appeal.TeamID == nil {
			slog.Warn(fmt.Sprintf("matcher: appeal has no team id: %d", appeal.ID))
			continue
		}

		candidates := byTeam[*appeal.TeamID]
		if len(candidates) == 0 {
			continue
		}

		best := pickBestManager(candidates, freeSlots)
		if best == nil {
			continue
		}

		slotID := pickOldestFreeSlot(best, freeSlots)
		if slotID == 0 {
			continue
		}

		best.activeAppealsCount++
		t := now
		best.lastAssign = &t
		best.usedSlots[slotID] = struct{}{}

		out = append(out, AssignPayload{
			AppealID:  appeal.ID,
			ManagerID: best.emploee.ID,
			SlotID:    slotID,
			TeamID:    *appeal.TeamID,
			Priority:  m.classifyAppealPriority(appeal),
		})
	}

	return out
}

func pickBestManager(candidates []*employeeState, freeSlots map[int][]model.Slot) *employeeState {
	var best *employeeState
	for _, cur := range candidates {
		if best == nil {
			if hasFreeSlot(cur, freeSlots) {
				best = cur
			}
			continue
		}
		if !hasFreeSlot(cur, freeSlots) {
			continue
		}
		if cur.activeAppealsCount != best.activeAppealsCount {
			if cur.activeAppealsCount < best.activeAppealsCount {
				best = cur
			}
			continue
		}
		curLast, bestLast := time.Time{}, time.Time{}
		if cur.lastAssign != nil {
			curLast = *cur.lastAssign
		}
		if best.lastAssign != nil {
			bestLast = *best.lastAssign
		}
		if !curLast.Equal(bestLast) {
			if curLast.Before(bestLast) {
				best = cur
			}
			continue
		}
		if oldestFreeSlotTime(cur, freeSlots).Before(oldestFreeSlotTime(best, freeSlots)) {
			best = cur
			continue
		}
		// if countFreeSlots(cur, freeSlots) > countFreeSlots(best, freeSlots) {
		// 	best = cur
		// }
	}
	return best
}

func hasFreeSlot(m *employeeState, freeSlots map[int][]model.Slot) bool {
	for _, s := range freeSlots[m.emploee.ID] {
		if _, used := m.usedSlots[s.ID]; !used {
			return true
		}
	}
	return false
}

func pickOldestFreeSlot(m *employeeState, freeSlots map[int][]model.Slot) int {
	for _, s := range freeSlots[m.emploee.ID] {
		if _, used := m.usedSlots[s.ID]; used {
			continue
		}
		return s.ID
	}
	return 0
}

func oldestFreeSlotTime(m *employeeState, freeSlots map[int][]model.Slot) time.Time {
	for _, s := range freeSlots[m.emploee.ID] {
		if _, used := m.usedSlots[s.ID]; used {
			continue
		}
		if s.UpdatedAt != nil {
			return *s.UpdatedAt
		}
		return time.Time{}
	}
	return time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
}

func countFreeSlots(m *employeeState, freeSlots map[int][]model.Slot) int {
	n := 0
	for _, s := range freeSlots[m.emploee.ID] {
		if _, used := m.usedSlots[s.ID]; !used {
			n++
		}
	}
	return n
}

func (m *Matcher) classifyAppealPriority(a model.Appeal) int {
	importance := m.appealService.IsImportant(a.ID)
	if importance {
		return 0
	}
	return 1
}
