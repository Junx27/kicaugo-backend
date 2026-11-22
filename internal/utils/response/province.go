package response

// ProvinceData adalah representasi province di response.
type ProvinceData struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProvinceListOutput adalah output untuk GET list.
type ProvinceListOutput struct {
	Status int `status:"200"`
	Body   Success[[]ProvinceData]
}

// scaffold:response-types
