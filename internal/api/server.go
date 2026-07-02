package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ajaykumarsingh/flow/internal/focus"
	"github.com/ajaykumarsingh/flow/internal/journal"
	"github.com/ajaykumarsingh/flow/internal/mood"
	"github.com/ajaykumarsingh/flow/internal/store"
	"github.com/ajaykumarsingh/flow/internal/task"
)

type Server struct {
	taskSvc    *task.Service
	moodSvc    *mood.Service
	journalSvc *journal.Service
	focusSvc   *focus.Service
	apiKey     string
}

func NewServer(db *store.DB, apiKey string) *Server {
	return &Server{
		taskSvc:    task.NewService(db),
		moodSvc:    mood.NewService(db),
		journalSvc: journal.NewService(db),
		focusSvc:   focus.NewService(db),
		apiKey:     apiKey,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/openapi.json", s.handleOpenAPI)
	mux.HandleFunc("/health", s.handleHealth)

	// auth-gated routes
	mux.HandleFunc("/context", s.auth(s.handleContext))
	mux.HandleFunc("/tasks", s.auth(s.handleTasks))
	mux.HandleFunc("/tasks/suggest", s.auth(s.handleSuggest))
	mux.HandleFunc("/tasks/", s.auth(s.handleTaskByID))
	mux.HandleFunc("/checkins", s.auth(s.handleCheckins))
	mux.HandleFunc("/notes", s.auth(s.handleNotes))
	mux.HandleFunc("/insights", s.auth(s.handleInsights))

	return corsMiddleware(mux)
}

func (s *Server) Listen(addr string) error {
	log.Printf("flow API server listening on %s", addr)
	log.Printf("OpenAPI spec: http://%s/openapi.json", addr)
	return http.ListenAndServe(addr, s.Handler())
}

// ── middleware ────────────────────────────────────────────────────────────────

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" {
			next(w, r)
			return
		}
		key := r.Header.Get("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")
		if key != s.apiKey {
			writeError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next(w, r)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeBody(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
