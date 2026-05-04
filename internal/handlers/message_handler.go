package handlers

import (
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/Ikit24/gomini/internal/database"
)

func (s *Server) HandleCreateMessage(w http.ResponseWriter, r *http.Request) {
	type messageParams struct {
		Content string `json:"content"`
	}

	var params messageParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	sessionIDString := r.PathValue("id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session id format")
		return
	}

	userMessage := database.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    session.UserID,
		Role:      database.UserRole,
		Content:   params.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.DB.CreateMessage(&userMessage)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't create message")
		return
	}

	aiResponse, err := s.AI.GenerateContent(r.Context(), params.Content)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to get response from the AI")
		return
	}

	aiMessage := database.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    session.UserID,
		Role:      database.ModelRole,
		Content:   aiResponse,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.DB.CreateMessage(&aiMessage)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't create message")
		return
	}

	RespondWithJSON(w, http.StatusCreated, aiMessage)
}

func (s *Server) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session id format")
		return
	}

	messages, err := s.DB.GetMessagesBySessionID(sessionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't retrieve messages")
		return
	}

	RespondWithJSON(w, http.StatusOK, messages)
}
