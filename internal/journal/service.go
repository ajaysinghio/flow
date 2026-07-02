package journal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/oklog/ulid/v2"
)

type Note struct {
	ID        string
	Content   string
	Tags      []string
	Timestamp time.Time
}

type Service struct {
	db *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Add(content string, tags []string) (*Note, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if tags == nil {
		tags = []string{}
	}
	tagsJSON, _ := json.Marshal(tags)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
	n := &Note{ID: id, Content: content, Tags: tags, Timestamp: time.Now()}
	_, err := s.db.Exec(
		`INSERT INTO notes (id, content, tags, timestamp) VALUES (?,?,?,?)`,
		n.ID, n.Content, string(tagsJSON), n.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("save note: %w", err)
	}
	return n, nil
}

func (s *Service) Recent(n int) ([]*Note, error) {
	rows, err := s.db.Query(
		`SELECT id, content, tags, timestamp FROM notes ORDER BY timestamp DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string
		if err := rows.Scan(&note.ID, &note.Content, &tagsJSON, &note.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		out = append(out, &note)
	}
	return out, rows.Err()
}
