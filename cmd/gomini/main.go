package main

import (
	"os"
	"log"
	"fmt"
	"context"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/Ikit24/gomini/internal/handlers"
	"github.com/google/uuid"
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
		log.Fatal(err)
	}
	defer db.Close()

	u := &database.User{Name: "Ati"}
	err = db.CreateUser(u)
	if err != nil {
		fmt.Println("couldn't create user:", err)
		return
	}
	fmt.Printf("Created user with ID: %w\n", u.ID)

	foundUser, err := db.GetUserByName("Ati")
	if err != nil {
		log.Fatal(err)
	}

	if foundUser != nil {
		fmt.Printf("Found user in DB: %s (ID: %d)\n", foundUser.Name, foundUser.ID)
	} else {
		fmt.Println("User not found!")
	}

	sessionID := uuid.New()
	msg := &database.Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   "Hello! Is this message being saved?",
	}

	if err := db.SaveMessage(msg); err != nil {
		log.Fatal("Couldn't save message:", err)
	}

	fmt.Printf("Successfully save message! (ID: %s)\n", msg.ID)

	history, err := db.GetMessagesBySessionID(sessionID)
	if err == nil {
		fmt.Printf("Thread history (%d messages):\n", len(history))
		for _, m := range history {
			fmt.Printf("[%s]: %s\n", m.Role, m.Content)
		}
	}

	fmt.Printf("Cleaning up session...")
	err = db.DeleteSession(sessionID)
	if err != nil {
		log.Fatal(err)
	}

	finalHistory, _ := db.GetMessagesBySessionID(sessionID)
	fmt.Printf("History count after deletion: %d\n", len(finalHistory))
}
