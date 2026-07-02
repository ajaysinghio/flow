package mood

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/oklog/ulid/v2"
)

type CheckIn struct {
	ID        string
	Mood      int // 1–5
	Energy    int // 1–5
	Note      string
	Timestamp time.Time
}

type Service struct {
	db *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Save(mood, energy int, note string) (*CheckIn, error) {
	if mood < 1 || mood > 5 {
		return nil, fmt.Errorf("mood must be 1–5")
	}
	if energy < 1 || energy > 5 {
		return nil, fmt.Errorf("energy must be 1–5")
	}
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
	c := &CheckIn{ID: id, Mood: mood, Energy: energy, Note: note, Timestamp: time.Now()}
	_, err := s.db.Exec(
		`INSERT INTO checkins (id, mood, energy, note, timestamp) VALUES (?,?,?,?,?)`,
		c.ID, c.Mood, c.Energy, c.Note, c.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("save checkin: %w", err)
	}
	return c, nil
}

// Latest returns the most recent check-in within maxAge. Returns nil if none.
func (s *Service) Latest(maxAge time.Duration) (*CheckIn, error) {
	cutoff := time.Now().Add(-maxAge)
	row := s.db.QueryRow(
		`SELECT id, mood, energy, note, timestamp FROM checkins WHERE timestamp > ? ORDER BY timestamp DESC LIMIT 1`,
		cutoff,
	)
	var c CheckIn
	err := row.Scan(&c.ID, &c.Mood, &c.Energy, &c.Note, &c.Timestamp)
	if err != nil {
		return nil, nil // no recent check-in is fine
	}
	return &c, nil
}

// Recent returns the last n check-ins ordered newest first.
func (s *Service) Recent(n int) ([]*CheckIn, error) {
	rows, err := s.db.Query(
		`SELECT id, mood, energy, note, timestamp FROM checkins ORDER BY timestamp DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*CheckIn
	for rows.Next() {
		var c CheckIn
		if err := rows.Scan(&c.ID, &c.Mood, &c.Energy, &c.Note, &c.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}
