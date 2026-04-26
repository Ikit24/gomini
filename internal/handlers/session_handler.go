package handlers

import (
	"net/http"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

func (s *Server) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	type sessionParams struct {
		Name string `json:"name"`
		UserID string `json:"user_id"`
	}

	var params sessionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	sessionToCreate := database.Session{
		UserID: userID,
		Title: params.Name,
	}

	err := s.DB.CreateSession(&sessionToCreate)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error couldn't create session")
		return
	}

	RespondWithJSON(w, http.StatusCreated, sessionToCreate)
}

func (s *Server) HandleGetSessionByUserID(w http.ResponseWriter, r *http.Request) {
	type sessionByUserID struct {
		UserID string `json:"user_id"`
	}

	var params sessionByUserID
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	sessions, err := s.DB.GetSessionsByUserID(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error couldn't get session")
		return
	}

	RespondWithJSON(w, http.StatusOK, sessions)
}
