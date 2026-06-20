package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/google/uuid"
)

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userToCreate := database.User{
		ID:        uuid.New(),
		Name:      params.Name,
		Email:     params.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.DB.CreateUser(&userToCreate)
	if err != nil {
		log.Printf("Database error: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "error couldn't create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, userToCreate)
}
