package response

// AuthData adalah data response untuk register dan login.
type AuthData struct {
	User  UserData `json:"user"`
	Token string   `json:"token"`
}

// UserData adalah representasi user di response.
type UserData struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// RegisterOutput adalah output untuk registrasi.
type RegisterOutput struct {
	Status int `status:"201"`
	Body   Success[AuthData]
}

// LoginOutput adalah output untuk login.
type LoginOutput struct {
	Status int `status:"200"`
	Body   Success[AuthData]
}

// LogoutOutput adalah output untuk logout.
type LogoutOutput struct {
	Status int `status:"204"`
}
