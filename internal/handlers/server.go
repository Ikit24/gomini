package handlers

import (
	"net/http"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

type Server struct {
	DB *database.DB
	AI *gemini.Client
}

func NewServer(db *database.DB, ai *gemini.Client) *Server {
	return &Server{
		DB: db,
		AI: ai,
	}
}

func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", s.HandleCreateUser)
	
	mux.HandleFunc("GET /api/users/{user_id}/sessions", s.HandleGetSessionByUserID)
	mux.HandleFunc("POST /api/users/{user_id}/sessions", s.HandleCreateSession)

	mux.HandleFunc("POST /api/sessions/{id}/messages", s.HandleCreateMessage)

	return http.ListenAndServe(addr, mux)
}
