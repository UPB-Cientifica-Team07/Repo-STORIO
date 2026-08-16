package repository

import (
	"errors"
	"sync"
)

type User struct {
	ID       string
	Username string
	Email    string
	Password string
}

type UserRepository struct {
	mu           sync.RWMutex
	usersByID    map[string]*User
	usersByEmail map[string]*User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		usersByID:    make(map[string]*User),
		usersByEmail: make(map[string]*User),
	}
}

func (r *UserRepository) Create(
	user *User,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByEmail[user.Email]; exists {
		return errors.New("el usuario ya existe")
	}

	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user

	return nil
}

func (r *UserRepository) FindByEmail(
	email string,
) (*User, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByEmail[email]

	if !exists {
		return nil, errors.New("usuario no encontrado")
	}

	return user, nil
}

func (r *UserRepository) FindByID(
	id string,
) (*User, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByID[id]

	if !exists {
		return nil, errors.New("usuario no encontrado")
	}

	return user, nil
}
