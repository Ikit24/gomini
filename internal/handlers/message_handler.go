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
		sessionID: SessionID,
		Role:      database.RoleUser,
		Content:   params.Content,
	}

	err = s.DB.CreateMessage(&userMessage)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "couldn't create message")
		return
	}
}
