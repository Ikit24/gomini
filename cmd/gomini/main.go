package main

import (
	"log"
	"fmt"

	"github.com/Ikit24/gomini/internal/database"
)

func main() {
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
	fmt.Printf("Created user with ID: %d\n", u.ID)
}
