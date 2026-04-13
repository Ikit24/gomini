package internal

import {
	"github.com/Ikit24/gomini/internal/database"
}

type Server struct {
	DB *database.DB
	AI *gemini.Client
}
