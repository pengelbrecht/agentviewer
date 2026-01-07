package main

import (
	"fmt"
	"log"
)

// User represents a user in the system.
type User struct {
	ID   int
	Name string
	Age  int
}

// GetUser retrieves a user by ID.
func GetUser(id int) *User {
	// TODO: implement database lookup
	return &User{
		ID:   id,
		Name: "Unknown",
		Age:  0,
	}
}

// PrintUser prints user information.
func PrintUser(u *User) {
	fmt.Printf("User: %s (ID: %d)\n", u.Name, u.ID)
}

func main() {
	log.Println("Starting application...")

	user := GetUser(1)
	PrintUser(user)

	log.Println("Done.")
}
