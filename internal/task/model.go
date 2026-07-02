package task

import "time"

type Size string

const (
	SizeXS Size = "xs"
	SizeS  Size = "s"
	SizeM  Size = "m"
	SizeL  Size = "l"
	SizeXL Size = "xl"
)

type Energy string

const (
	EnergyLow  Energy = "low"
	EnergyMed  Energy = "med"
	EnergyHigh Energy = "high"
)

type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

type Task struct {
	ID          string
	Title       string
	Size        Size
	Energy      Energy
	Status      Status
	ParentID    *string
	Tags        []string
	DueDate     *time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// IsOverdue returns true if the task has a due date in the past.
func (t *Task) IsOverdue() bool {
	return t.DueDate != nil && t.DueDate.Before(time.Now())
}

// DueSoon returns true if due within 24 hours.
func (t *Task) DueSoon() bool {
	return t.DueDate != nil && !t.IsOverdue() && t.DueDate.Before(time.Now().Add(24*time.Hour))
}

// energyScore maps energy level to a numeric value for comparison.
func energyScore(e Energy) int {
	switch e {
	case EnergyLow:
		return 1
	case EnergyMed:
		return 2
	case EnergyHigh:
		return 3
	}
	return 2
}

// sizeScore maps size to points — smaller = higher score when energy is low.
func sizeScore(s Size) int {
	switch s {
	case SizeXS:
		return 5
	case SizeS:
		return 4
	case SizeM:
		return 3
	case SizeL:
		return 2
	case SizeXL:
		return 1
	}
	return 3
}
