package app

import "github.com/gtarcea/1DevDayTalk2014/schema"

// Objects implements the UsersService provide an technology
// independent way of managing the applications users.
type UsersService interface {
	CreateUser(email, fullname string) (schema.User, error)
	GetAll() ([]schema.User, error)
	GetUserByEmail(email string) (schema.User, error)
	UpdateUserByEmail(email, fullname string) (schema.User, error)
}

// The UserDBService provides a way for tracking and storing users.
// It provides no persistence guarantees, though specific
// implementations might provide persistence.
type UserDBService interface {
	GetByEmail(email string) (schema.User, error)
	GetAll() ([]schema.User, error)
	Insert(email, fullname string) (schema.User, error)
	Update(email, fullname string) (schema.User, error)
}
