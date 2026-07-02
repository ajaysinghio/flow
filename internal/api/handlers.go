package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/ajaykumarsingh/flow/internal/task"
)

// GET /context
func (s *Server) handleContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	tasks, _ := s.taskSvc.List(false)
	checkin, _ := s.moodSvc.Latest(4 * time.Hour)
	notes, _ := s.journalSvc.Recent(5)
	writeJSON(w, http.StatusOK, map[string]any{
		"tasks":          tasks,
		"latest_checkin": checkin,
		"recent_notes":   notes,
	})
}

// GET /tasks        — list pending tasks
// POST /tasks       — add a task
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		all := r.URL.Query().Get("all") == "true"
		tasks, err := s.taskSvc.List(all)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var body struct {
			Title    string   `json:"title"`
			Size     string   `json:"size"`
			Energy   string   `json:"energy"`
			Tags     []string `json:"tags"`
			ParentID *string  `json:"parent_id"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if body.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		if body.Size == "" {
			body.Size = "m"
		}
		if body.Energy == "" {
			body.Energy = "med"
		}
		t, err := s.taskSvc.Add(body.Title, task.Size(body.Size), task.Energy(body.Energy), body.Tags, body.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)

	default:
		writeError(w, http.StatusMethodNotAllowed, "use GET or POST")
	}
}

// GET  /tasks/suggest              — suggest one task
// PUT  /tasks/{id}/complete        — complete a task
// POST /tasks/{id}/breakdown       — break into steps
func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}

	// /tasks/suggest is handled here because mux matches /tasks/ prefix
	if id == "suggest" {
		s.handleSuggest(w, r)
		return
	}

	switch action {
	case "complete":
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "use PUT")
			return
		}
		t, err := s.taskSvc.Get(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		if err := s.taskSvc.Complete(t.ID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "done", "title": t.Title})

	case "breakdown":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "use POST")
			return
		}
		var body struct {
			Steps []string `json:"steps"`
		}
		if err := decodeBody(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		parent, err := s.taskSvc.Get(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		var created []*task.Task
		for _, step := range body.Steps {
			sub, err := s.taskSvc.Add(step, task.SizeXS, parent.Energy, nil, &parent.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			created = append(created, sub)
		}
		writeJSON(w, http.StatusCreated, map[string]any{"parent": parent.Title, "subtasks": created})

	default:
		writeError(w, http.StatusNotFound, "unknown action — try /complete or /breakdown")
	}
}

// GET /tasks/suggest
func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	checkin, _ := s.moodSvc.Latest(4 * time.Hour)
	energy := 3
	if checkin != nil {
		energy = checkin.Energy
	}
	tasks, _ := s.taskSvc.List(false)
	suggested := task.Suggest(tasks, energy)
	if suggested == nil {
		writeJSON(w, http.StatusOK, map[string]any{"task": nil, "message": "queue is empty or no tasks match current energy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task": suggested, "based_on_energy": energy})
}

// POST /checkins
func (s *Server) handleCheckins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Mood   int    `json:"mood"`
		Energy int    `json:"energy"`
		Note   string `json:"note"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	c, err := s.moodSvc.Save(body.Mood, body.Energy, body.Note)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// POST /notes
func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	n, err := s.journalSvc.Add(body.Content, body.Tags)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

// GET /insights
func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	days := 7
	checkins, _ := s.moodSvc.Recent(days * 3)
	tasks, _ := s.taskSvc.List(true)

	var moodSum, energySum int
	for _, c := range checkins {
		moodSum += c.Mood
		energySum += c.Energy
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var done, total int
	for _, t := range tasks {
		if t.CreatedAt.After(cutoff) {
			total++
			if t.Status == task.StatusDone {
				done++
			}
		}
	}

	result := map[string]any{
		"period_days":     days,
		"checkin_count":   len(checkins),
		"tasks_total":     total,
		"tasks_completed": done,
	}
	if len(checkins) > 0 {
		result["avg_mood"] = float64(moodSum) / float64(len(checkins))
		result["avg_energy"] = float64(energySum) / float64(len(checkins))
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /openapi.json — served without auth so AI tools can discover the spec
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpec))
}
