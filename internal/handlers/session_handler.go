package handlers

import (
	"errors"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
)

func (s *Server) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	type sessionParams struct {
		Name string `json:"name"`
	}

	var params sessionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userIDString := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	sessionToCreate := database.Session{
		UserID: userID,
		Title: params.Name,
	}

	err = s.DB.CreateSession(&sessionToCreate)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error couldn't create session")
		return
	}

	RespondWithJSON(w, http.StatusCreated, sessionToCreate)
}

func (s *Server) HandleGetSessionByUserID(w http.ResponseWriter, r *http.Request) {
	userIDString := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDString)
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

func (s *Server) HandleGetSessionByID(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session_id format")
		return
	}

	session, err := s.DB.GetSessionByID(sessionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error couldn't get session")
		return
	}

	RespondWithJSON(w, http.StatusOK, session)
}

func (s *Server) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	userIDString := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session_id format")
		return
	}

	err = s.DB.DeleteSession(sessionID, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "session not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, nil)
}

func (s *Server) HandleUpdateSession(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session_id format")
		return
	}

	type sessionParams struct {
		Title string `json:"title"`
	}

	var params sessionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = s.DB.UpdateSession(sessionID, params.Title)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't update session")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"status":"updated"})
}

func (s *Server) HandleListAllSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.DB.GetAllSessions()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't retrieve session list")
		return
	}

	RespondWithJSON(w, http.StatusOK, sessions)
}

func (s *Server) HandleDeleteSessionByID(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(userIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session_id format")
		return
	}

	err = s.DB.DeleteSessionBySessionID(sessionID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "session_id not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, nil)
}
