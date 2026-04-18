package main

import (
	"os"
	"log"
	"fmt"
	"context"
	"net/http"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	geminiKey := os.Getenv("GEMINI_API_KEY")
		if geminiKey == "" {
			log.Fatal("GEMINI_API_KEY missing")
		}

	ctx := context.Background()
	
	aiClient, err := gemini.NewClient(ctx, geminiKey)
	if err != nil {
		log.Fatal("couldn't initialize gemini client", err)
	}

	const dbPath = "gomini.db"
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal("couldn't open database", err)
	}
	defer db.Close()

	servr := handlers.NewServer(db, aiClient)
	log.Fatal(servr.ListenAndServ())
}
