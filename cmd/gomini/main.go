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

	foundUser, err := db.GetUserByName("Ati")
	if err != nil {
		log.Fatal(err)
	}

	if foundUser != nil {
		fmt.Printf("Found user in DB: %s (ID: %d)\n", foundUser.Name, foundUser.ID)
	} else {
		fmt.Println("User not found!")
	}
}
