// Package app defines the core entities and interfaces for user management.
package app

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrUserNotFound is returned when a requested user does not exist in the repository.
var ErrUserNotFound = errors.New("user not found")

// User represents an authenticated entity within the system.
type User struct {
	Username     string
	PasswordHash string
}

// NewUser creates a new user instance with a hashed password.
func NewUser(username, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	return User{Username: username, PasswordHash: string(hash)}, nil
}

// VerifyPassword checks if the provided plain-text password matches the stored hash.
func (u User) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// UserRepository defines the interface for persisting and retrieving user data.
type UserRepository interface {
	Add(user User) error
	FindByID(id string) (User, error)
}
