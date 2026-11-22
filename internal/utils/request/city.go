// Package request berisi tipe input request yang dipakai handler.
package request

// ListCitiesInput adalah input untuk daftar kabupaten/kota pada satu provinsi.
type ListCitiesInput struct {
	ProvinceID int64  `json:"province_id" query:"province_id" required:"true" minimum:"1" doc:"ID provinsi Kemendagri"`
	Query      string `json:"q" query:"q" maxLength:"100" doc:"Pencarian nama kabupaten/kota"`
}
