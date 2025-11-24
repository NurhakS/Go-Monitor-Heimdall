package repository

import (
	"database/sql"
	"uptime-monitor/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user record in the database
func (r *UserRepository) CreateUser(user models.User) error {
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, user.Username, user.Email, user.Password)
	return err
}

// GetUserByID retrieves a user by its ID
func (r *UserRepository) GetUserByID(id int) (models.User, error) {
	query := "SELECT id, username, email, password FROM users WHERE id = $1"
	row := r.db.QueryRow(query, id)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password); err != nil {
		return user, err
	}

	return user, nil
}
