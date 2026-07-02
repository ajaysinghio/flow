package task

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/oklog/ulid/v2"
)

type Service struct {
	db *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{db: db}
}

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
}

func (s *Service) Add(title string, size Size, energy Energy, tags []string, parentID *string) (*Task, error) {
	return s.AddWithDue(title, size, energy, tags, parentID, nil)
}

func (s *Service) AddWithDue(title string, size Size, energy Energy, tags []string, parentID *string, dueDate *time.Time) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if size == "" {
		size = SizeM
	}
	if energy == "" {
		energy = EnergyMed
	}
	if tags == nil {
		tags = []string{}
	}
	tagsJSON, _ := json.Marshal(tags)
	t := &Task{
		ID:        newID(),
		Title:     title,
		Size:      size,
		Energy:    energy,
		Status:    StatusTodo,
		ParentID:  parentID,
		Tags:      tags,
		DueDate:   dueDate,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(
		`INSERT INTO tasks (id, title, size, energy, status, parent_id, tags, due_date, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Title, t.Size, t.Energy, t.Status, t.ParentID, string(tagsJSON), t.DueDate, t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	return t, nil
}

func (s *Service) Complete(id string) error {
	res, err := s.db.Exec(
		`UPDATE tasks SET status='done', completed_at=datetime('now') WHERE id=? AND status != 'done'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task not found or already done")
	}
	return nil
}

func (s *Service) SetDoing(id string) error {
	_, err := s.db.Exec(`UPDATE tasks SET status='doing' WHERE id=? AND status='todo'`, id)
	return err
}

func (s *Service) List(includeAll bool) ([]*Task, error) {
	query := `SELECT id, title, size, energy, status, parent_id, tags, due_date, created_at, completed_at
	          FROM tasks WHERE status != 'done'`
	if includeAll {
		query = `SELECT id, title, size, energy, status, parent_id, tags, due_date, created_at, completed_at FROM tasks`
	}
	query += ` ORDER BY created_at ASC`
	return s.scan(query)
}

func (s *Service) Get(id string) (*Task, error) {
	rows, err := s.scan(
		`SELECT id, title, size, energy, status, parent_id, tags, due_date, created_at, completed_at FROM tasks WHERE id=?`, id,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return rows[0], nil
}

func (s *Service) Subtasks(parentID string) ([]*Task, error) {
	return s.scan(
		`SELECT id, title, size, energy, status, parent_id, tags, due_date, created_at, completed_at
		 FROM tasks WHERE parent_id=? ORDER BY created_at ASC`, parentID,
	)
}

var timeFormats = []string{
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseTime(s string) time.Time {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t := parseTime(*s)
	if t.IsZero() {
		return nil
	}
	return &t
}

func (s *Service) scan(query string, args ...any) ([]*Task, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []*Task
	for rows.Next() {
		var t Task
		var tagsJSON string
		var parentID *string
		var dueDateStr *string
		var completedAtStr *string
		var createdAtStr string
		if err := rows.Scan(&t.ID, &t.Title, &t.Size, &t.Energy, &t.Status,
			&parentID, &tagsJSON, &dueDateStr, &createdAtStr, &completedAtStr); err != nil {
			return nil, err
		}
		t.ParentID = parentID
		t.CreatedAt = parseTime(createdAtStr)
		t.DueDate = parseTimePtr(dueDateStr)
		t.CompletedAt = parseTimePtr(completedAtStr)
		_ = json.Unmarshal([]byte(tagsJSON), &t.Tags)
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}
