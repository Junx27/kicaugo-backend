package request

// CreateOrganizerInput adalah input POST /api/v1/organizers.
type CreateOrganizerInput struct {
	Body struct {
		Name        string  `json:"name" required:"true" minLength:"3" maxLength:"120" doc:"Nama penyelenggara"`
		Slug        string  `json:"slug" required:"true" minLength:"3" maxLength:"120" pattern:"^[a-z0-9]+(?:-[a-z0-9]+)*$" doc:"Slug unik penyelenggara"`
		Email       string  `json:"email" required:"true" format:"email" maxLength:"255" doc:"Email penyelenggara"`
		Phone       *string `json:"phone,omitempty" maxLength:"30" doc:"Nomor telepon"`
		LogoURL     *string `json:"logo_url,omitempty" maxLength:"500" doc:"URL logo"`
		Description *string `json:"description,omitempty" maxLength:"1000" doc:"Deskripsi penyelenggara"`
	}
}

// RegisterOrganizerMemberInput adalah input POST /api/v1/organizers/{id}/members.
type RegisterOrganizerMemberInput struct {
	ID   int64 `path:"id" minimum:"1" doc:"ID organizer"`
	Body struct {
		Role string `json:"role,omitempty" enum:"owner,admin,staff,verifier" default:"staff" doc:"Peran anggota organizer"`
	}
}
