package handlers

import (
	"net/http"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
)

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userToCreate := database.User{
		Email: params.Email,
		Name: params.Name,
	}

	err := s.DB.CreateUser(&userToCreate)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error couldn't create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, userToCreate)
}


