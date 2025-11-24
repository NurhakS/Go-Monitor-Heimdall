// Package models defines the data structures and database operations for users
package models

import (
	"errors" // Import the errors package for error handling
)

// User represents a user in the system
type User struct {
	ID    string `json:"id"`    // Unique identifier for the user
	Name  string `json:"name"`  // Name of the user
	Email string `json:"email"` // Email of the user
}

// In-memory user storage (for demonstration purposes)
var users = []User{}

// GetAllUsers retrieves all users from the in-memory storage
func GetAllUsers() ([]User, error) {
	return users, nil // Return the list of users
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id string) (User, error) {
	for _, user := range users {
		if user.ID == id {
			return user, nil // Return the found user
		}
	}
	return User{}, errors.New("user not found") // Return error if not found
}

// CreateUser adds a new user to the in-memory storage
func CreateUser(user *User) error {
	users = append(users, *user) // Add the new user to the list
	return nil                   // Return no error
}

// UpdateUser updates an existing user by their ID
func UpdateUser(id string, updatedUser *User) error {
	for i, user := range users {
		if user.ID == id {
			users[i] = *updatedUser // Update the user
			return nil              // Return no error
		}
	}
	return errors.New("user not found") // Return error if not found
}

// DeleteUser removes a user by their ID
func DeleteUser(id string) error {
	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...) // Remove the user from the list
			return nil                                // Return no error
		}
	}
	return errors.New("user not found") // Return error if not found
}
