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
		params.Role: RoleUser,
	}

	messageToCreate := s.AI
	
	// type Message struct {
	// 	ID        uuid.UUID `json:"id"         db:"id"`
	// 	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	// 	Role      RoleType  `json:"role"       db:"role"`
	// 	Content   string    `json:"content"    db:"content"`
	// 	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// }

}
