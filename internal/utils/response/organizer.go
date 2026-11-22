package response

// OrganizerData adalah representasi organizer di response.
type OrganizerData struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone"`
	LogoURL     *string `json:"logo_url"`
	Description *string `json:"description"`
	IsVerified  bool    `json:"is_verified"`
}

// OrganizerMemberData adalah representasi anggota organizer di response.
type OrganizerMemberData struct {
	ID          int64  `json:"id"`
	OrganizerID int64  `json:"organizer_id"`
	UserID      int64  `json:"user_id"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
}

// CreateOrganizerOutput adalah output POST /api/v1/organizers.
type CreateOrganizerOutput struct {
	Status int `status:"201"`
	Body   Success[OrganizerData]
}

// RegisterOrganizerMemberOutput adalah output POST /api/v1/organizers/{id}/members.
type RegisterOrganizerMemberOutput struct {
	Status int `status:"201"`
	Body   Success[OrganizerMemberData]
}
