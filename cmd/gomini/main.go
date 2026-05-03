package main

import (
	"os"
	"os/signal"
	"net/http"
	"log"
	"context"
	"time"

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
	defer aiClient.Close()

	const dbPath = "gomini.db"
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal("couldn't open database", err)
	}
	defer db.Close()

	servr := handlers.NewServer(db, aiClient)

	go func() {
		log.Println("🚀 Server starting on http://localhost:8080")
		if err := servr.ListenAndServe(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10 * time.Second)
	defer cancel()

	err = servr.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
