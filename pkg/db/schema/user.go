package schema

import (
	"time"
)

// User models a user in the system.
type User struct {
	ID          string    `gorethink:"id,omitempty" json:"id"`
	Email       string    `gorethink:"email" json:"email"`
	Admin       bool      `gorethink:"admin" json:"admin"`
	Fullname    string    `gorethink:"fullname" json:"fullname"`
	Password    string    `gorethink:"password" json:"-"`
	APIKey      string    `gorethink:"apikey" json:"-"`
	Birthtime   time.Time `gorethink:"birthtime" json:"birthtime"`
	MTime       time.Time `gorethink:"mtime" json:"mtime"`
	Avatar      string    `gorethink:"avatar" json:"avatar"`
	Description string    `gorethink:"description" json:"description"`
	Affiliation string    `gorethink:"affiliation" json:"affiliation"`
	HomePage    string    `gorethink:"homepage" json:"homepage"`
	Type        string    `gorethink:"_type" json:"_type"`
}

// NewUser creates a new User instance.
func NewUser(name, email, password, apikey string) User {
	now := time.Now()
	return User{
		ID:        email,
		Fullname:  name,
		Email:     email,
		Password:  password,
		APIKey:    apikey,
		Birthtime: now,
		MTime:     now,
		Type:      "user",
	}
}
