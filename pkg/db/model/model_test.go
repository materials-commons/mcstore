package model

import (
	"fmt"
	"testing"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcdstore/pkg/db/schema"
)

var _ = fmt.Println

var (
	session, _ = r.Connect(
		r.ConnectOpts{
			Address:  "localhost:30815",
			Database: "mctestdb",
		})
)

func TestGetUserModel(t *testing.T) {
	m := &rModel{
		schema: schema.User{},
		table:  "users",
	}

	var user schema.User
	err := m.Qs(session).ByID("test@mc.org", &user)
	if err != nil {
		t.Errorf("Lookup by Id failed: %s", err)
	}

	if user.ID != "test@mc.org" {
		t.Errorf("Unexpected user return %#v", user)
	}

	var users []schema.User
	err = m.Qs(session).Rows(m.Table(), &users)
	if err != nil {
		t.Errorf("Lookup all users failed: %s", err)
	}

	if len(users) == 0 {
		t.Errorf("No users returned when looking up all users")
	}
}

func TestInsertDeleteUserModel(t *testing.T) {
	m := Users
	u := schema.NewUser("tuser", "tuser@test.org", "abc123", "apikey123")
	var user schema.User
	err := m.Qs(session).Insert(u, &user)
	if err != nil {
		t.Fatalf("Not able to insert new user: %s", err)
	}

	err = m.Qs(session).Delete(user.ID)
	if err != nil {
		t.Fatalf("Unable to delete id %s: %s", user.ID, err)
	}

	err = m.Qs(session).Insert(u, user)
	if err != mcerr.ErrInvalid {
		t.Fatalf("Passed in wrong type and did not get error")
	}

	err = m.Qs(session).Insert(u, nil)
	if err != nil {
		t.Fatalf("Performed insert without retrieving value and got error: %s", err)
	}
	err = m.Qs(session).Delete(user.ID)
	if err != nil {
		t.Errorf("Unable to delete user with id %s: %s", user.ID, err)
	}
}

func TestUpdateDeleteUserModel(t *testing.T) {
	var u schema.User
	err := Users.Qs(session).ByID("test@mc.org", &u)
	if err != nil {
		t.Fatalf("Unable to retrieve test@mc.org: %s", err)
	}

	u.Description = "change1"
	err = Users.Qs(session).Update(u.ID, u)
	if err != nil {
		t.Fatalf("Unable to update user test@mc.org: %s", err)
	}
	var u2 schema.User
	Users.Qs(session).ByID("test@mc.org", &u2)
	if u2.Description != "change1" {
		t.Errorf("Description not updated, expected 'change1', got '%s'", u2.Description)
	}

	// Test with map
	err = Users.Qs(session).Update(u.ID, map[string]interface{}{"description": "change2"})
	Users.Qs(session).ByID("test@mc.org", &u2)
	if u2.Description != "change2" {
		t.Errorf("Description not updated, expected 'change2', got '%s'", u2.Description)
	}
}

func TestGetRows(t *testing.T) {
	var users []schema.User
	rql := r.Table("users")
	err := GetRows(rql, session, &users)
	if err != nil {
		t.Errorf("GetRows all users failed: %s", err)
	}

	if len(users) == 0 {
		t.Errorf("Users length == 0")
	}

	err = GetRows(rql, session, users)
	if err == nil {
		t.Errorf("Unexpected nil error when passing in bad parameter")
	}
}
