package services

import "testing"

var testUService = NewUsers(NewUsersDB())

func TestCreateRetrieveUser(t *testing.T) {
	u, err := testUService.CreateUser("test@test.com", "Test Me")
	if err != nil {
		t.Fatalf("Unable to insert: %s", err)
	}

	if u.Email != "test@test.com" || u.Fullname != "Test Me" {
		t.Fatalf("CreateUser returned record with unexpected values: %#v", u)
	}

	u, err = testUService.GetUserByEmail("test@test.com")
	if err != nil {
		t.Fatalf("Unable to lookup user by email: %s", err)
	}

	if u.Email != "test@test.com" || u.Fullname != "Test Me" {
		t.Fatalf("Lookup returned wrong record: %#v", u)
	}
}
