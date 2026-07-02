package focus

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/oklog/ulid/v2"
)

type Session struct {
	ID          string
	TaskID      *string
	StartedAt   time.Time
	EndedAt     *time.Time
	Interrupted bool
}

type Service struct {
	db *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Start(taskID *string) (*Session, error) {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
	sess := &Session{ID: id, TaskID: taskID, StartedAt: time.Now()}
	_, err := s.db.Exec(
		`INSERT INTO focus_sessions (id, task_id, started_at) VALUES (?,?,?)`,
		sess.ID, sess.TaskID, sess.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}
	return sess, nil
}

func (s *Service) End(id string, interrupted bool) error {
	_, err := s.db.Exec(
		`UPDATE focus_sessions SET ended_at=datetime('now'), interrupted=? WHERE id=?`,
		interrupted, id,
	)
	return err
}
