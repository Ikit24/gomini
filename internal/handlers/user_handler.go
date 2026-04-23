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
	mux.HandleFunc("POST /api/sessions", s.HandleCreateSession)
	mux.HandleFunc("POST /api/sessions/{id}/messages", s.HandleCreateSession)
	mux.HandleFunc("GET /api/sessions/{id}", s.HandleGetSessionByUserID)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userToCreate := database.User{
		Email: params.Email,
		Name: params.Name,
	}

	err := s.DB.CreateUser(&userToCreate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, userToCreate)
}

func (s *Server) HandleCreateSession(w http.ResponseWriter, r *http.Request) {}

func (s *Server) HandleGetSessionByUserID(w http.ResponseWriter, r *http.Request) {}
