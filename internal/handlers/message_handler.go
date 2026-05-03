package handlers

import (
	"net/http"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

func (s *Server) HandlerCreateMessage(w http.ResponseWriter, r *http.Request) {
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
		SessionID: sessionID,
		Role:      database.RoleUser,
		Content:   params.Content,
	}

	err = s.DB.CreateMessage(&userMessage)
	if err != nil {
		RespondWithError(w, http.InternalServerError, "couldn't create message")
		return
	}

	aiResponse, err := s.AI.GenerateContent(r.Context(), params.Content)
	if err != nil {
		RespondWithError(w, http.InternalServerError, "failed to get response from the AI")
		return
	}

	aiMessage := database.Message{
		SessionID: sessionID,
		Role:      database.RoleAssistant,
		Content:   aiResponse,
	}

	err = s.DB.CreateMessage(&aiMessage)
	if err != nil {
		RespondWithError(w, http.InternalServerError, "couldn't create message")
		return
	}

	RespondWithJSON(w, http.StatusCreated, aiMessage)
}

func (s *Server) HandlerListMessages(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session id format")
		return
	}

	messages, err := s.DB.GetMessagesBySessionID(sessionID)
	if err != nil {
		RespondWithError(w, http.InternalServerError, "couldn't retrieve messages")
		return
	}

	RespondWithJSON(w, http.StatusOK, messages)
}
