package entities

// AuthResult adalah hasil autentikasi yang dikirim ke klien.
type AuthResult struct {
	User  User
	Token string
}
