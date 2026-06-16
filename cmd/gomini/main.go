package main

import (
	"os"
	"net/http"
	"fmt"
	"log"
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Ikit24/gomini/internal/tui"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/database"
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
	defer aiClient.Close()

	const dbPath = "gomini.db"
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal("couldn't open database", err)
	}
	defer db.Close()

	servr := handlers.NewServer(db, aiClient)

	user, err := db.GetUserByName("ati")
	if err != nil {
		log.Fatalf("Failed to query database for user: %v", err)
	}

	if user == nil {
		user = &database.User{
			Name: "ati",
			Email: "ati@local.dev",
		}
		err = db.CreateUser(user)
		if err != nil {
			log.Fatalf("Failed to bootstrap local user: %v", err)
		}
	}

	sessions, err := db.GetSessionsByUserID(user.ID)
	if err != nil {
		log.Fatalf("Faield to fetch past sessions: %v", err)
	}

	go func() {
		log.Println("🚀 Server starting on http://localhost:8080")
		if err := servr.ListenAndServe(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	p:= tea.NewProgram(tui.InitialModel(db, aiClient, user.ID, sessions))
	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
		os.Exit(1)
	}

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10 * time.Second)
	defer cancel()

	err = servr.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
