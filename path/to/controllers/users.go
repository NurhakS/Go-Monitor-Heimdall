// Package controllers provides HTTP handlers for user-related operations
package controllers

import (
	"encoding/json"       // Import the encoding/json package for JSON handling
	"net/http"            // Import the net/http package for HTTP handling
	"your_project/models" // Import the models package for User model

	"github.com/gorilla/mux" // Import the Gorilla Mux router
)

// GetAllUsers handles the request to fetch all users
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := models.GetAllUsers() // Fetch all users from the database
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // Handle errors
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set response header
	json.NewEncoder(w).Encode(users)                   // Send the users as a JSON response
}

// GetUserByID handles the request to fetch a user by ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                     // Get URL variables
	userID := vars["id"]                    // Extract user ID from URL
	user, err := models.GetUserByID(userID) // Find user by ID
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound) // Handle user not found
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set response header
	json.NewEncoder(w).Encode(user)                    // Send the user as a JSON response
}

// CreateUser handles the request to create a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User                         // Create a new user instance
	err := json.NewDecoder(r.Body).Decode(&user) // Decode request body into user
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Handle bad request
		return
	}
	err = models.CreateUser(&user) // Save the user to the database
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // Handle errors
		return
	}
	w.WriteHeader(http.StatusCreated) // Send created status
	json.NewEncoder(w).Encode(user)   // Send the created user as a JSON response
}

// UpdateUser handles the request to update a user by ID
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                          // Get URL variables
	userID := vars["id"]                         // Extract user ID from URL
	var user models.User                         // Create a new user instance
	err := json.NewDecoder(r.Body).Decode(&user) // Decode request body into user
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Handle bad request
		return
	}
	err = models.UpdateUser(userID, &user) // Update user in the database
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound) // Handle user not found
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set response header
	json.NewEncoder(w).Encode(user)                    // Send the updated user as a JSON response
}

// DeleteUser handles the request to delete a user by ID
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)              // Get URL variables
	userID := vars["id"]             // Extract user ID from URL
	err := models.DeleteUser(userID) // Delete user by ID
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound) // Handle user not found
		return
	}
	w.WriteHeader(http.StatusNoContent) // Send no content response
}
