// Package app provides the core business logic and repository implementations.
package app

import "sync"

// UserInMemoryRepo implements a thread-safe in-memory user repository.
type UserInMemoryRepo struct {
	mu    sync.RWMutex
	users map[string]User
}

// NewUserInMemoryRepo initializes and returns a new instance of UserInMemoryRepo.
func NewUserInMemoryRepo() *UserInMemoryRepo {
	return &UserInMemoryRepo{users: make(map[string]User)}
}

// Add inserts a new user into the in-memory repository.
func (r *UserInMemoryRepo) Add(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.Username] = user
	return nil
}

// FindByID retrieves a user by their username from the repository.
func (r *UserInMemoryRepo) FindByID(id string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}
