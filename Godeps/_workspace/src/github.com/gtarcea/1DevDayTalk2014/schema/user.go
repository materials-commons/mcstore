package schema

// User describes an individual user in the system.
type User struct {
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
}
