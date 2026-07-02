package task

import (
	"sort"
	"time"
)

// Suggest returns the single best task for the given energy level (1–5).
// Never returns a list — one answer, always.
func Suggest(tasks []*Task, energyLevel int) *Task {
	eligible := filterByEnergy(tasks, energyLevel)
	if len(eligible) == 0 {
		return nil
	}
	scored := scoreAll(eligible, energyLevel)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	return scored[0].task
}

type scored struct {
	task  *Task
	score float64
}

func filterByEnergy(tasks []*Task, currentEnergy int) []*Task {
	maxEnergy := energyThreshold(currentEnergy)
	var out []*Task
	for _, t := range tasks {
		if t.Status == StatusDone {
			continue
		}
		if energyScore(t.Energy) <= maxEnergy {
			out = append(out, t)
		}
	}
	return out
}

// energyThreshold maps current energy 1–5 to max task energy score.
// Low energy (1–2) → only low tasks. Medium (3) → low+med. High (4–5) → all.
func energyThreshold(e int) int {
	if e <= 2 {
		return 1 // low only
	}
	if e <= 3 {
		return 2 // low + med
	}
	return 3 // all
}

func scoreAll(tasks []*Task, energyLevel int) []scored {
	now := time.Now()
	results := make([]scored, len(tasks))
	for i, t := range tasks {
		results[i] = scored{task: t, score: scoreTask(t, energyLevel, now)}
	}
	return results
}

func scoreTask(t *Task, energyLevel int, now time.Time) float64 {
	var score float64

	// Size: prefer smaller tasks when energy is low
	if energyLevel <= 2 {
		score += float64(sizeScore(t.Size)) * 2.0
	} else {
		score += float64(sizeScore(t.Size)) * 0.5
	}

	// Age bonus: older tasks get a nudge to prevent infinite deferral
	ageDays := now.Sub(t.CreatedAt).Hours() / 24
	score += min(ageDays*0.3, 3.0)

	// Continuity: already in-progress tasks get a strong boost
	if t.Status == StatusDoing {
		score += 4.0
	}

	return score
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
