// Package routes defines the application routes and their handlers
package routes

import (
	"your_project/controllers" // Import the controllers package

	"github.com/gorilla/mux" // Import the Gorilla Mux router
)

// InitializeRoutes sets up the application routes
func InitializeRoutes() *mux.Router {
	router := mux.NewRouter() // Create a new router
	// Define routes and associate them with controller functions
	router.HandleFunc("/users", controllers.GetAllUsers).Methods("GET")        // Route to get all users
	router.HandleFunc("/users/{id}", controllers.GetUserByID).Methods("GET")   // Route to get user by ID
	router.HandleFunc("/users", controllers.CreateUser).Methods("POST")        // Route to create a new user
	router.HandleFunc("/users/{id}", controllers.UpdateUser).Methods("PUT")    // Route to update user by ID
	router.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE") // Route to delete user by ID
	return router                                                              // Return the configured router
}
