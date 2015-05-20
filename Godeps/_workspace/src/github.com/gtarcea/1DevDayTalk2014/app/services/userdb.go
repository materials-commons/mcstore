package services

import (
	"sync"

	"github.com/gtarcea/1DevDayTalk2014/app"
	"github.com/gtarcea/1DevDayTalk2014/schema"
)

// usersDB implements an in memory db store. It synchronizes
// access to the underlying map.
type usersDB struct {
	users map[string]schema.User
	mutex sync.RWMutex
}

// NewUsersDB creates a new instance of the usersDB.
func NewUsersDB() app.UserDBService {
	db := &usersDB{
		users: make(map[string]schema.User),
	}
	// Insert one example user for demo purposes
	db.Insert("example@user.com", "example user")
	return db
}

// GetByEmail looks up a user by their email address. It returns
// ErrNotFound if the email could not be found.
func (db *usersDB) GetByEmail(email string) (schema.User, error) {
	defer db.mutex.RUnlock()
	db.mutex.RLock()

	user, ok := db.users[email]
	if !ok {
		return schema.User{}, app.ErrNotFound
	}

	return user, nil
}

// GetAll returns all known users. It will return an empty list
// if there are no users.
func (db *usersDB) GetAll() ([]schema.User, error) {
	defer db.mutex.RUnlock()
	db.mutex.RLock()

	users := make([]schema.User, 0, len(db.users))
	i := 0
	for _, user := range db.users {
		users = append(users, user)
		i++
	}
	return users, nil
}

// Insert adds a new user to the system.
func (db *usersDB) Insert(email, fullname string) (schema.User, error) {
	defer db.mutex.Unlock()
	db.mutex.Lock()

	u := schema.User{
		Email:    email,
		Fullname: fullname,
	}
	db.users[email] = u
	return u, nil
}

// Update updates and existing user
func (db *usersDB) Update(email, fullname string) (schema.User, error) {
	defer db.mutex.Unlock()
	db.mutex.Lock()

	user, ok := db.users[email]
	if !ok {
		return schema.User{}, app.ErrNotFound
	}

	user.Fullname = fullname
	return user, nil
}
