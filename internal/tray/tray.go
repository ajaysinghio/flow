package tray

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/ajaykumarsingh/flow/internal/app"
	"github.com/ajaykumarsingh/flow/internal/task"
)

type state struct {
	mu          sync.Mutex
	currentTask *task.Task
}

// Run starts the systray app. Must be called from the main goroutine.
func Run(a *app.App) {
	s := &state{}
	systray.Run(func() { onReady(a, s) }, func() {})
}

func onReady(a *app.App, s *state) {
	systray.SetIcon(makeIcon())
	systray.SetTooltip("flow")

	// current task display (disabled — informational only)
	mTask := systray.AddMenuItem("Loading…", "Your next task")
	mTask.Disable()

	mDone := systray.AddMenuItem("✓  Mark done", "Complete current task")
	mNext := systray.AddMenuItem("↻  Refresh", "Refresh suggestion")

	systray.AddSeparator()

	mAdd := systray.AddMenuItem("+ Add task…", "Capture a new task")

	// check-in submenu
	mCheckin := systray.AddMenuItem("◎  Check in", "Log your current energy")
	energyItems := []string{
		"1  drained", "2  low", "3  medium", "4  good", "5  charged",
	}
	mEnergy := make([]*systray.MenuItem, len(energyItems))
	for i, label := range energyItems {
		mEnergy[i] = mCheckin.AddSubMenuItem(label, fmt.Sprintf("Set energy to %d/5", i+1))
	}

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit flow", "")

	// fan energy clicks into a single channel
	energyCh := make(chan int, 5)
	for i := range mEnergy {
		go func(idx int) {
			for range mEnergy[idx].ClickedCh {
				energyCh <- idx + 1
			}
		}(i)
	}

	// initial load
	refresh(a, s, mTask, mDone)

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				refresh(a, s, mTask, mDone)

			case <-mDone.ClickedCh:
				s.mu.Lock()
				t := s.currentTask
				s.mu.Unlock()
				if t != nil {
					_ = a.Tasks.Complete(t.ID)
					refresh(a, s, mTask, mDone)
				}

			case <-mNext.ClickedCh:
				refresh(a, s, mTask, mDone)

			case <-mAdd.ClickedCh:
				title, ok := inputDialog("What do you need to do?", "")
				if ok && title != "" {
					_, _ = a.Tasks.Add(title, task.SizeM, task.EnergyMed, nil, nil)
					refresh(a, s, mTask, mDone)
				}

			case level := <-energyCh:
				_, _ = a.Moods.Save(level, level, "")
				refresh(a, s, mTask, mDone)

			case <-mQuit.ClickedCh:
				ticker.Stop()
				systray.Quit()
			}
		}
	}()
}

func refresh(a *app.App, s *state, mTask *systray.MenuItem, mDone *systray.MenuItem) {
	checkin, _ := a.Moods.Latest(4 * time.Hour)
	energy := 3
	if checkin != nil {
		energy = checkin.Energy
	}

	tasks, _ := a.Tasks.List(false)
	suggested := task.Suggest(tasks, energy)

	s.mu.Lock()
	s.currentTask = suggested
	s.mu.Unlock()

	if suggested == nil {
		mTask.SetTitle("✓ Queue empty")
		mDone.Disable()
		systray.SetTooltip("flow — nothing to do!")
	} else {
		title := suggested.Title
		if len(title) > 40 {
			title = title[:37] + "…"
		}
		mTask.SetTitle("→  " + title)
		mDone.Enable()
		systray.SetTooltip(fmt.Sprintf("flow — %s", suggested.Title))
	}
}
