package response

// BirdCategoryPayload adalah representasi kategori burung di response.
type BirdCategoryPayload struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

// BirdCategoryListOutput adalah output untuk GET /bird-categories.
type BirdCategoryListOutput struct {
	Status int `status:"200"`
	Body   Success[[]BirdCategoryPayload]
}

// BirdcategoryDeleteOutput adalah output untuk DELETE.
type BirdcategoryDeleteOutput struct {
	Status int `status:"204"`
}

// scaffold:response-types
