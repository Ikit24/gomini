package handlers

import (
	"net/http"
	"context"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

type Server struct {
	DB *database.DB
	AI *gemini.Client
	httpServer *http.Server
}

func NewServer(db *database.DB, ai *gemini.Client) *Server {
	return &Server{
		DB: db,
		AI: ai,
	}
}

func (s *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	err := s.DB.Ping()
	if err != nil {
		RespondWithError(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"status":"available"})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", s.HandleCreateUser)

	mux.HandleFunc("POST /api/users/{user_id}/sessions", s.HandleCreateSession)
	mux.HandleFunc("GET /api/users/{user_id}/sessions", s.HandleGetSessionByUserID)

	mux.HandleFunc("GET /healthz", s.HandleHealthCheck)
	mux.HandleFunc("GET /api/sessions/{id}", s.HandleGetSessionByID)
	mux.HandleFunc("PATCH /api/sessions/{id}", s.HandleUpdateSession)
	mux.HandleFunc("DELETE /api/users/{user_id}/sessions/{session_id}", s.HandleDeleteSession)

	mux.HandleFunc("POST /api/sessions/{session_id}/messages", s.HandleCreateMessage)
	mux.HandleFunc("GET /api/sessions/{session_id}/messages", s.HandleListMessages)

	var handler http.Handler = mux
	handler = s.middleware(handler)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return s.httpServer.ListenAndServe()
}
