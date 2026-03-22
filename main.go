package main

import (
	"log"

	"github.com/Ikit24/gomini/internal/database"
)

func main() {
	const dbPath = "gomini.db"
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
