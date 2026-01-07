package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRepository handles user data operations.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUser retrieves a user by ID from the database.
func (r *UserRepository) GetUser(ctx context.Context, id int) (*User, error) {
	query := `SELECT id, name, email, age, created_at FROM users WHERE id = $1`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Age,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d: %w", id, err)
	}

	return &user, nil
}

// PrintUser prints user information to the console.
func PrintUser(u *User) {
	fmt.Printf("User: %s <%s> (ID: %d)\n", u.Name, u.Email, u.ID)
	fmt.Printf("  Age: %d\n", u.Age)
	fmt.Printf("  Created: %s\n", u.CreatedAt.Format(time.RFC3339))
}

func main() {
	log.Println("Starting application...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize database connection
	db, err := sql.Open("postgres", "postgres://localhost/users?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	user, err := repo.GetUser(ctx, 1)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	PrintUser(user)

	log.Println("Done.")
}
