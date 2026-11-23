package format

import (
	"fmt"
	"strconv"
	"strings"
)

// Rupiah memformat angka menjadi string mata uang Rupiah, contoh: Rp 2.000.000.
func Rupiah(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	s := strconv.FormatInt(amount, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return fmt.Sprintf("Rp %s%s", sign, b.String())
}

// RupiahString memformat string angka menjadi Rupiah. Jika bukan angka, mengembalikan nilai asli dengan prefix Rp.
func RupiahString(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return "-"
	}
	// hapus pemisah umum
	clean := strings.ReplaceAll(strings.ReplaceAll(v, ".", ""), ",", "")
	n, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return "Rp " + v
	}
	return Rupiah(n)
}
