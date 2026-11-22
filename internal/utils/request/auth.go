// Package request berisi tipe input request yang dipakai handler.
package request

// RegisterInput adalah input untuk registrasi.
type RegisterInput struct {
	Body struct {
		Name       string `json:"name" required:"true" minLength:"3" maxLength:"100" doc:"Nama pengguna"`
		Identifier string `json:"identifier" required:"true" format:"email" doc:"Email atau nomor WhatsApp"`
		Password   string `json:"password" required:"true" minLength:"8" maxLength:"72" doc:"Kata sandi"`
	}
}

// LoginInput adalah input untuk login.
type LoginInput struct {
	Body struct {
		Identifier string `json:"identifier" required:"true" doc:"Email atau nomor WhatsApp"`
		Password   string `json:"password" required:"true" doc:"Kata sandi"`
	}
}
