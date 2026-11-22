// Package request berisi tipe input request yang dipakai handler.
package request

// ListCitiesInput adalah input untuk daftar kabupaten/kota pada satu provinsi.
type ListCitiesInput struct {
	ProvinceID int64  `query:"province_id" required:"true" minimum:"1" doc:"ID provinsi Kemendagri"`
	Name       string `query:"name" maxLength:"100" doc:"Pencarian nama kabupaten/kota"`
}
