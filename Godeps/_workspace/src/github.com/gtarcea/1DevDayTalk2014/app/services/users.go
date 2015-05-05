package services

import (
	"github.com/gtarcea/1DevDayTalk2014/app"
	"github.com/gtarcea/1DevDayTalk2014/schema"
)

// usersService provides an implementation of app.UsersService.
type usersService struct {
	db app.UserDBService
}

// NewUsers creates a new usersService.
func NewUsers(db app.UserDBService) app.UsersService {
	return &usersService{
		db: db,
	}
}

// CreateUser creates a new user.
func (u *usersService) CreateUser(email, fullname string) (schema.User, error) {
	return u.db.Insert(email, fullname)
}

// GetAll returns a list of all users.
func (u *usersService) GetAll() ([]schema.User, error) {
	return u.db.GetAll()
}

// GetUserByEmail looks up a user by email.
func (u *usersService) GetUserByEmail(email string) (schema.User, error) {
	return u.db.GetByEmail(email)
}

// UpdateUserByEmail locates a user by email address and updates their fullname
func (u *usersService) UpdateUserByEmail(email, fullname string) (schema.User, error) {
	return u.db.Update(email, fullname)
}
