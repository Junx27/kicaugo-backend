package response

// BirdCategoryData adalah representasi kategori burung di response.
type BirdCategoryData struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

// BirdCategoryListOutput adalah output untuk GET /bird-categories.
type BirdCategoryListOutput struct {
	Status int `status:"200"`
	Body   Success[[]BirdCategoryData]
}

// scaffold:response-types
