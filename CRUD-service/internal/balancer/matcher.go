package balancer

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/hibiken/asynq"
)

type managerState struct {
	m          ManagerRow
	active     int
	lastAssign *time.Time
	usedSlots  map[string]struct{}
}

type Matcher struct {
	db    *DB
	asynq *asynq.Client
	cfg   Config
}

func NewMatcher(db *DB, asynqClient *asynq.Client, cfg Config) *Matcher {
	return &Matcher{db: db, asynq: asynqClient, cfg: cfg}
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
	appeals, err := m.db.FetchPendingAppeals(ctx, m.cfg.FetchAppealsLimit)
	if err != nil {
		return err
	}
	managers, err := m.db.FetchAvailableManagers(ctx, m.cfg.FetchManagersLimit)
	if err != nil {
		return err
	}

	if len(appeals) == 0 || len(managers) == 0 {
		return nil
	}

	managerIDs := make([]string, 0, len(managers))
	for _, mg := range managers {
		managerIDs = append(managerIDs, mg.ID)
	}
	freeSlots, err := m.db.FetchFreeSlotsByManagers(ctx, managerIDs)
	if err != nil {
		return err
	}

	assignments := FindOptimalAssignments(appeals, managers, freeSlots)
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

// FindOptimalAssignments is a simplified port of MatchService:
// - appeals are already ordered by DB priority, but we still group/sort deterministically
// - managers are chosen by active appeals, last assign, oldest free slot, free slots count
func FindOptimalAssignments(appeals []AppealRow, managers []ManagerRow, freeSlots map[string][]SlotRow) []AssignPayload {
	// Index managers by team.
	byTeam := make(map[string][]*managerState)
	states := make(map[string]*managerState, len(managers))
	for _, mg := range managers {
		st := &managerState{
			m:         mg,
			active:    mg.ActiveAppeals,
			lastAssign: mg.LastAssignAt,
			usedSlots:  map[string]struct{}{},
		}
		states[mg.ID] = st
		for _, team := range mg.TeamIDs {
			byTeam[team] = append(byTeam[team], st)
		}
	}

	// Stable sort managers per team by baseline criteria.
	for team, arr := range byTeam {
		sort.SliceStable(arr, func(i, j int) bool {
			a, b := arr[i], arr[j]
			if a.active != b.active {
				return a.active < b.active
			}
			ai := time.Time{}
			bi := time.Time{}
			if a.lastAssign != nil {
				ai = *a.lastAssign
			}
			if b.lastAssign != nil {
				bi = *b.lastAssign
			}
			if !ai.Equal(bi) {
				return ai.Before(bi)
			}
			return len(freeSlots[a.m.ID]) > len(freeSlots[b.m.ID])
		})
		byTeam[team] = arr
	}

	now := time.Now().UTC()
	out := make([]AssignPayload, 0)

	for _, ap := range appeals {
		candidates := byTeam[ap.TeamID]
		if len(candidates) == 0 {
			continue
		}

		best := pickBestManager(candidates, freeSlots)
		if best == nil {
			continue
		}

		slotID := pickOldestFreeSlot(best, freeSlots)
		if slotID == "" {
			continue
		}

		best.active++
		t := now
		best.lastAssign = &t
		best.usedSlots[slotID] = struct{}{}

		out = append(out, AssignPayload{
			AppealID:  ap.ID,
			ManagerID: best.m.ID,
			SlotID:    slotID,
			TeamID:    ap.TeamID,
			Priority:  classifyAppealPriority(ap),
		})
	}

	return out
}

func pickBestManager(candidates []*managerState, freeSlots map[string][]SlotRow) *managerState {
	var best *managerState
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
		// Compare criteria similar to TS.
		if cur.active != best.active {
			if cur.active < best.active {
				best = cur
			}
			continue
		}
		curLast := time.Time{}
		bestLast := time.Time{}
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
		curOldest := oldestFreeSlotTime(cur, freeSlots)
		bestOldest := oldestFreeSlotTime(best, freeSlots)
		if curOldest.Before(bestOldest) {
			best = cur
			continue
		}
		if countFreeSlots(cur, freeSlots) > countFreeSlots(best, freeSlots) {
			best = cur
		}
	}
	return best
}

func hasFreeSlot(m *managerState, freeSlots map[string][]SlotRow) bool {
	for _, s := range freeSlots[m.m.ID] {
		if _, used := m.usedSlots[s.ID]; !used {
			return true
		}
	}
	return false
}

func pickOldestFreeSlot(m *managerState, freeSlots map[string][]SlotRow) string {
	for _, s := range freeSlots[m.m.ID] {
		if _, used := m.usedSlots[s.ID]; used {
			continue
		}
		return s.ID // slots are ordered by updated_at ASC from DB
	}
	return ""
}

func oldestFreeSlotTime(m *managerState, freeSlots map[string][]SlotRow) time.Time {
	for _, s := range freeSlots[m.m.ID] {
		if _, used := m.usedSlots[s.ID]; used {
			continue
		}
		return s.UpdatedAt
	}
	return time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
}

func countFreeSlots(m *managerState, freeSlots map[string][]SlotRow) int {
	n := 0
	for _, s := range freeSlots[m.m.ID] {
		if _, used := m.usedSlots[s.ID]; !used {
			n++
		}
	}
	return n
}

func classifyAppealPriority(a AppealRow) int {
	if a.IsImportant && a.IsUrgent {
		return 10
	}
	if a.IsImportant {
		return 8
	}
	if a.IsUrgent {
		return 6
	}
	return 5
}

