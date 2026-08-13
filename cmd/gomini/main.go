package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"flag"
	"path/filepath"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/handlers"
	"github.com/Ikit24/gomini/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Try to find the OS-specific config folder
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Could not find user config directory: %v", err)
	}

	// 2. Define the app's config directory and ensure it exists
	appConfigDir := filepath.Join(configDir, "gomini")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// 3. Construct path for the .env file and attempt to load it
	envPath := filepath.Join(appConfigDir, ".env")
	_ = godotenv.Load(envPath)

	// 4. Now grab the key, whether it came from the .env file or the system environment
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("GEMINI_API_KEY missing. Please set it in your environment or ~/.config/gomini/.env")
	}

	ctx := context.Background()

	var fileFlag = flag.String("file", "", "path to a file to attach as a context")
	flag.Parse()

	var fileContent string
	if *fileFlag != "" {
		data, err := os.ReadFile(*fileFlag)
		if err != nil {
			log.Fatal("couldn't open file", err)
		}
		fileContent = string(data)
	}

	aiClient, err := gemini.NewClient(ctx, geminiKey, fileContent)
	if err != nil {
		log.Fatal("couldn't initialize gemini client", err)
	}

	dbPath := filepath.Join(appConfigDir, "gomini.db")
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal("couldn't open database", err)
	}
	defer db.Close()

	if err := db.PurgeEmptyMessages(); err != nil {
		log.Printf("warning: Failed to auto-purge empty messages from the database: %v", err)
	}

	servr := handlers.NewServer(db, aiClient)

	user, err := db.GetUserByName("ati")
	if err != nil {
		log.Fatalf("failed to query database for user: %v", err)
	}

	if user == nil {
		user = &database.User{
			Name:  "ati",
			Email: "ati@local.dev",
		}
		err = db.CreateUser(user)
		if err != nil {
			log.Fatalf("failed to bootstrap local user: %v", err)
		}
	}

	sessions, err := db.GetSessionsByUserID(user.ID)
	if err != nil {
		log.Fatalf("failed to fetch past sessions: %v", err)
	}

	go func() {
		log.Println("🚀 Server running on http://localhost:8080")
		if err := servr.ListenAndServe(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	p := tea.NewProgram(
		tui.InitialModel(db, aiClient, user.ID, sessions),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
		os.Exit(1)
	}
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = servr.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
