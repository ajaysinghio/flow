package tray

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
)

const maxPickItems = 5

type state struct {
	mu          sync.Mutex
	currentTask *task.Task
	ranked      []*task.Task
}

// Run starts the systray app. Must be called from the main goroutine.
func Run(a *app.App) {
	s := &state{}
	systray.Run(func() { onReady(a, s) }, func() {})
}

func onReady(a *app.App, s *state) {
	systray.SetIcon(makeIcon())
	systray.SetTooltip("flow")

	// suggested task display (disabled — informational only)
	mTask := systray.AddMenuItem("Loading…", "Your suggested task")
	mTask.Disable()
	mDone := systray.AddMenuItem("✓  Mark done", "Complete this task")
	mNext := systray.AddMenuItem("↻  Refresh", "Refresh suggestion")

	systray.AddSeparator()

	// pick submenu — top 5 tasks
	mPick := systray.AddMenuItem("≡  Pick a task", "Choose from your queue")
	pickItems := make([]*systray.MenuItem, maxPickItems)
	for i := range pickItems {
		pickItems[i] = mPick.AddSubMenuItem("—", "")
		pickItems[i].Disable()
	}

	mAdd := systray.AddMenuItem("+  Add task…", "Capture a new task")

	// check-in submenu
	mCheckin := systray.AddMenuItem("◎  Check in", "Log your current energy")
	energyLabels := []string{"1  drained", "2  low", "3  medium", "4  good", "5  charged"}
	mEnergy := make([]*systray.MenuItem, len(energyLabels))
	for i, label := range energyLabels {
		mEnergy[i] = mCheckin.AddSubMenuItem(label, fmt.Sprintf("Set energy to %d/5", i+1))
	}

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit flow", "")

	// fan energy clicks into one channel
	energyCh := make(chan int, 5)
	for i := range mEnergy {
		go func(idx int) {
			for range mEnergy[idx].ClickedCh {
				energyCh <- idx + 1
			}
		}(i)
	}

	// fan pick item clicks into one channel carrying the index
	pickCh := make(chan int, maxPickItems)
	for i := range pickItems {
		go func(idx int) {
			for range pickItems[idx].ClickedCh {
				pickCh <- idx
			}
		}(i)
	}

	refresh(a, s, mTask, mDone, pickItems)

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				refresh(a, s, mTask, mDone, pickItems)

			case <-mDone.ClickedCh:
				s.mu.Lock()
				t := s.currentTask
				s.mu.Unlock()
				if t != nil {
					_ = a.Tasks.Complete(t.ID)
					refresh(a, s, mTask, mDone, pickItems)
				}

			case <-mNext.ClickedCh:
				refresh(a, s, mTask, mDone, pickItems)

			case idx := <-pickCh:
				s.mu.Lock()
				ranked := s.ranked
				s.mu.Unlock()
				if idx < len(ranked) {
					chosen := ranked[idx]
					_ = a.Tasks.SetDoing(chosen.ID)
					s.mu.Lock()
					s.currentTask = chosen
					s.mu.Unlock()
					title := chosen.Title
					if len(title) > 40 {
						title = title[:37] + "…"
					}
					mTask.SetTitle("→  " + title)
					mDone.Enable()
					systray.SetTooltip(fmt.Sprintf("flow — %s", chosen.Title))
				}

			case <-mAdd.ClickedCh:
				title, ok := inputDialog("What do you need to do?", "")
				if ok && title != "" {
					_, _ = a.Tasks.Add(title, task.SizeM, task.EnergyMed, nil, nil)
					refresh(a, s, mTask, mDone, pickItems)
				}

			case level := <-energyCh:
				_, _ = a.Moods.Save(level, level, "")
				refresh(a, s, mTask, mDone, pickItems)

			case <-mQuit.ClickedCh:
				ticker.Stop()
				systray.Quit()
			}
		}
	}()
}

func refresh(a *app.App, s *state, mTask *systray.MenuItem, mDone *systray.MenuItem, pickItems []*systray.MenuItem) {
	checkin, _ := a.Moods.Latest(4 * time.Hour)
	energy := 3
	if checkin != nil {
		energy = checkin.Energy
	}

	allTasks, _ := a.Tasks.List(false)
	ranked := task.Ranked(allTasks, energy)

	s.mu.Lock()
	s.ranked = ranked
	if len(ranked) > 0 {
		s.currentTask = ranked[0]
	} else {
		s.currentTask = nil
	}
	s.mu.Unlock()

	// update pick submenu
	for i, item := range pickItems {
		if i < len(ranked) {
			t := ranked[i]
			label := t.Title
			if len(label) > 36 {
				label = label[:33] + "…"
			}
			if d := task.FormatDue(t); d != "" {
				label += "  (" + d + ")"
			}
			item.SetTitle(label)
			item.Enable()
		} else {
			item.SetTitle("—")
			item.Disable()
		}
	}

	if len(ranked) == 0 {
		mTask.SetTitle("✓  Queue empty")
		mDone.Disable()
		systray.SetTooltip("flow — nothing to do!")
	} else {
		t := ranked[0]
		title := t.Title
		if len(title) > 40 {
			title = title[:37] + "…"
		}
		mTask.SetTitle("→  " + title)
		mDone.Enable()
		systray.SetTooltip(fmt.Sprintf("flow — %s", t.Title))
	}
}
