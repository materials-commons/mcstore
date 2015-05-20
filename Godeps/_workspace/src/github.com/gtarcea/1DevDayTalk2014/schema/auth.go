package schema

// Auth holds the authentication information for a user. It is
// passed back to a client when they successfully login.
type Auth struct {
	Username string `json:"username"` // The username
	Token    string `json:"token"`    // The JWT token
}
